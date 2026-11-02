// Windows local smoke test. Run after building AP2 frontend; requires Chrome and a multi-series package in runtime/temp.
// DICOM bytes stay on localhost. The native Go bridge is covered separately by go test ./apps/ap2-viewer.
import fs from 'node:fs';
import path from 'node:path';
import http from 'node:http';
import crypto from 'node:crypto';
import {spawn} from 'node:child_process';
const root=process.cwd(), dist=path.join(root,'apps/ap2-viewer/frontend/dist');
let study, content;
for(const name of fs.readdirSync(path.join(root,'runtime/temp'))) {
 const dir=path.join(root,'runtime/temp',name);
 try {
  const data=fs.readFileSync(path.join(dir,'study.dat'));
  const dec=crypto.createDecipheriv('aes-256-gcm',Buffer.from('12345678901234567890123456789012'),data.subarray(0,12));
  dec.setAuthTag(data.subarray(-16));
  const m=JSON.parse(Buffer.concat([dec.update(data.subarray(12,-16)),dec.final()]));
  m.series=m.series.filter(s=>s.modality?.trim().toUpperCase()!=='SR' && s.files?.length>0);
  m.series.forEach(s=>s.files.sort((a,b)=>a.position-b.position));
  if(m.series.length>1 && m.series[0].files.length>2) {study=m;content=dir;break;}
 } catch {}
}
if(!study) throw Error('No local multi-series study available');
const state={dataFound:true,manifest:study,imageCount:study.series.reduce((n,s)=>n+s.files.length,0)};
const mime={'.html':'text/html','.js':'text/javascript','.css':'text/css','.png':'image/png'};
const server=http.createServer((req,res)=>{
 try {
  const url=new URL(req.url,'http://localhost');
  if(url.pathname==='/state'){res.setHeader('Content-Type','application/json');res.end(JSON.stringify(state));return;}
  if(url.pathname==='/dicom'){
   const uid=url.searchParams.get('uid');if(!uid || /[\\/:]/.test(uid))throw Error('Invalid uid');
   let file=path.join(content,'data',uid);if(!fs.existsSync(file))file+='.dcm';
   res.end(fs.readFileSync(file).toString('base64'));return;
  }
  const file=path.resolve(dist,'.'+(url.pathname==='/'?'/index.html':url.pathname));
  if(!file.startsWith(dist+path.sep))throw Error('Invalid path');
  res.setHeader('Content-Type',mime[path.extname(file)]||'application/octet-stream');res.end(fs.readFileSync(file));
 }catch{res.statusCode=404;res.end('Not found');}
});
await new Promise(r=>server.listen(0,'127.0.0.1',r));
const port=server.address().port;
const browser=spawn('C:/Program Files/Google/Chrome/Application/chrome.exe',['--headless=new','--disable-gpu','--no-first-run','--remote-debugging-port=9338','--remote-allow-origins=*','--user-data-dir='+path.join(root,'runtime/ap2-browser-test'),'about:blank'],{windowsHide:true,stdio:'ignore'});
let socket;
try {
 let targets;
 for(let i=0;i<100;i++){try{targets=await(await fetch('http://127.0.0.1:9338/json')).json();break;}catch{await new Promise(r=>setTimeout(r,100));}}
 socket=new WebSocket(targets.find(t=>t.type==='page').webSocketDebuggerUrl);
 await new Promise(r=>socket.addEventListener('open',r,{once:true}));
 let seq=0;const pending=new Map(),exceptions=[];
 socket.addEventListener('message',({data})=>{const m=JSON.parse(data);if(m.id){const p=pending.get(m.id);pending.delete(m.id);m.error?p.reject(Error(m.error.message)):p.resolve(m.result);}else if(m.method==='Runtime.exceptionThrown') exceptions.push(m.params.exceptionDetails.text);});
 const call=(method,params={})=>new Promise((resolve,reject)=>{const id=++seq;pending.set(id,{resolve,reject});socket.send(JSON.stringify({id,method,params}));});
 const evaluate=async expression=>{const result=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(result.exceptionDetails)throw Error(result.exceptionDetails.text);return result.result.value;};
 await call('Runtime.enable');await call('Page.enable');
 await call('Emulation.setDeviceMetricsOverride',{width:1280,height:800,deviceScaleFactor:1,mobile:false});
 await call('Page.addScriptToEvaluateOnNewDocument',{source:`window.go={main:{App:{LoadStudy:()=>fetch('/state').then(r=>r.json()),GetDicomFile:uid=>fetch('/dicom?uid='+encodeURIComponent(uid)).then(r=>{if(!r.ok)throw Error('Read failed');return r.text()})}}};`});
 await call('Page.navigate',{url:'http://127.0.0.1:'+port});
 let rendered=false;
 for(let i=0;i<150;i++){
  rendered=await evaluate(`Boolean(window.cornerstone?.getEnabledElements().some(e=>e.image && e.element.getBoundingClientRect().left>=0))`);
  if(rendered)break;await new Promise(r=>setTimeout(r,200));
 }
 if(!rendered)throw Error('No DICOM image rendered; runtime exceptions: '+exceptions.join(','));
 await new Promise(r=>setTimeout(r,2500));
 const first=await evaluate(`(()=>{const e=window.cornerstone.getEnabledElements().find(e=>e.image&&e.element.getBoundingClientRect().left>=0);const r=e.element.getBoundingClientRect();return {id:e.image.imageId,width:e.image.width,height:e.image.height,x:r.x+r.width/2,y:r.y+r.height/2,stack:window.cornerstoneTools.getToolState(e.element,'stack').data[0].imageIds.length};})()`);
 await call('Input.dispatchMouseEvent',{type:'mouseMoved',x:first.x,y:first.y});
 await call('Input.dispatchMouseEvent',{type:'mouseWheel',x:first.x,y:first.y,deltaX:0,deltaY:130});
 await new Promise(r=>setTimeout(r,1200));
 const next=await evaluate(`(()=>{const e=window.cornerstone.getEnabledElements().find(e=>e.image&&e.element.getBoundingClientRect().left>=0);return {id:e.image.imageId,index:window.cornerstoneTools.getToolState(e.element,'stack').data[0].currentImageIdIndex};})()`);
 if(first.id===next.id || next.index===0)throw Error('Wheel did not advance the image');
 const pixels=await evaluate(`(()=>{const e=window.cornerstone.getEnabledElements().find(e=>e.image&&e.element.getBoundingClientRect().left>=0);const p=e.image.getPixelData();let min=Infinity,max=-Infinity;for(const v of p){min=Math.min(min,v);max=Math.max(max,v);}return {min,max};})()`);
 if(pixels.min===pixels.max)throw Error('Decoded image has no pixel variation');
 const thumbs=await evaluate(`Array.from(document.querySelectorAll('img[src^="data:image/png"]')).map(im=>{const c=document.createElement('canvas');c.width=im.naturalWidth;c.height=im.naturalHeight;const ctx=c.getContext('2d');ctx.drawImage(im,0,0);const p=ctx.getImageData(0,0,c.width,c.height).data;let min=255,max=0;for(let i=0;i<p.length;i+=4){min=Math.min(min,p[i]);max=Math.max(max,p[i]);}return {min,max};})`);
 if(!thumbs.some(t=>t.max>t.min))throw Error('All thumbnails are uniform');
 await evaluate(`document.querySelectorAll('[draggable="true"]')[1].click()`);
 await new Promise(r=>setTimeout(r,2500));
 const changed=await evaluate(`(()=>{const e=window.cornerstone.getEnabledElements().find(e=>e.image&&e.element.getBoundingClientRect().left>=0);return e?.image.imageId;})()`);
 if(!changed || changed===next.id)throw Error('Series selection did not change the image');
 console.log(JSON.stringify({seriesSelection:true,thumbnailPixels:true,pixelVariation:true,rendered:true,width:first.width,height:first.height,stackImages:first.stack,wheelNavigation:true,index:next.index,exceptions}));

 console.log('THUMBNAILS',await evaluate(`document.querySelectorAll('img[src^="data:image/png"]').length`));
 await call('Browser.close');
}finally{socket?.close();browser.kill();server.close();}