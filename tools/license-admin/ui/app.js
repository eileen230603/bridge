const $=id=>document.getElementById(id);
let licenses=[],filter='all',selected=null,toastTimer;
const labels={active:'Activa',soon:'Por vencer',expired:'Vencida',permanent:'Permanente'};
const escapeHTML=value=>{const el=document.createElement('span');el.textContent=value??'';return el.innerHTML.replaceAll(String.fromCharCode(34),"&quot;").replaceAll(String.fromCharCode(39),"&#39;")};
function state(row){if(row.permanent)return 'permanent';const left=Date.parse(row.expiresAt)-Date.now();return left<=0?'expired':left<=30*86400000?'soon':'active'}
function remaining(row){if(row.permanent)return 'Sin vencimiento';let ms=Date.parse(row.expiresAt)-Date.now();if(ms<=0)return 'Vencida';let minutes=Math.ceil(ms/60000);let days=Math.floor(minutes/1440),hours=Math.floor(minutes%1440/60);if(days)return `${days} día${days===1?'':'s'}${hours?` · ${hours} h`:''}`;if(hours)return `${hours} h · ${minutes%60} min`;return `${minutes} min`}
function date(value){return new Intl.DateTimeFormat('es-BO',{day:'2-digit',month:'short',year:'numeric',timeZone:'UTC'}).format(new Date(value))}
function notify(message){$('toast').textContent=message;$('toast').hidden=false;clearTimeout(toastTimer);toastTimer=setTimeout(()=>$('toast').hidden=true,3500)}
function fail(err){$('error').textContent=String(err?.message||err);$('error').hidden=false}
function api(){if(!window.go?.main?.Admin)throw new Error('No se pudo conectar con el administrador. Abre Symphony License Admin.exe.');return window.go.main.Admin}
function render(){
 let counts={all:licenses.length,active:0,soon:0,expired:0};for(const row of licenses){let s=state(row);if(s==='expired')counts.expired++;else counts.active++;if(s==='soon')counts.soon++}
 $('total').textContent=counts.all;$('active').textContent=counts.active;$('soon').textContent=counts.soon;$('expired').textContent=counts.expired;$('navCount').textContent=counts.all;
 document.querySelectorAll('.tabs [data-filter]').forEach(el=>{el.classList.toggle('selected',el.dataset.filter===filter);el.setAttribute('aria-pressed',String(el.dataset.filter===filter))});
 const q=$('search').value.trim().toLocaleLowerCase('es');
 const visible=licenses.filter(row=>{const s=state(row);return (filter==='all'||(filter==='active'?s!=='expired':s===filter))&&(!q||`${row.client} ${row.machineId}`.toLocaleLowerCase('es').includes(q))});
 $('rows').innerHTML=visible.map(row=>{let s=state(row);let initials=row.client.split(/\s+/).slice(0,2).map(w=>w[0]||'').join('').toUpperCase();return `<tr><td><div class="client-cell"><div class="client-avatar">${escapeHTML(initials)}</div><div><div class="client-name" title="${escapeHTML(row.client)}">${escapeHTML(row.client)}</div><div class="machine-id" title="${escapeHTML(row.machineId)}">${escapeHTML(row.machineId.slice(0,10))}…${escapeHTML(row.machineId.slice(-6))}</div></div></div></td><td><span class="badge ${s}">${labels[s]}</span></td><td><span class="remaining ${s}">${remaining(row)}</span><span class="subtext">${row.permanent?'Acceso permanente':s==='expired'?'Vigencia finalizada':'Actualizado automáticamente'}</span></td><td>${row.permanent?'—':date(row.expiresAt)}<span class="subtext">${row.permanent?'Sin fecha límite':'Fin del día · UTC'}</span></td><td><button class="row-action" data-detail="${row.id}">Ver licencia ↗</button></td></tr>`}).join('');
 $('empty').hidden=visible.length>0;if(!visible.length){$('empty').querySelector('h3').textContent=licenses.length?'No hay coincidencias':'Tu primera licencia empieza aquí';$('empty').querySelector('p').textContent=licenses.length?'Prueba con otro filtro, cliente o Machine ID.':'Agrega un cliente y su Machine ID para generar una licencia.';$('emptyNew').hidden=licenses.length>0}
 $('count').textContent=`${visible.length} de ${licenses.length} licencias`;
 if(selected && $('detailDialog').open)renderDetail();
}
async function load(){try{$('refresh').disabled=true;licenses=await api().ListLicenses()||[];$('error').hidden=true;render()}catch(err){fail(err)}finally{$('refresh').disabled=false}}
function openCreate(){$('createForm').reset();$('formError').hidden=true;const future=new Date();future.setUTCFullYear(future.getUTCFullYear()+1);$('expires').value=future.toISOString().slice(0,10);$('expires').min=new Date().toISOString().slice(0,10);$('dateLabel').hidden=false;$('expires').required=true;$('createDialog').showModal();$('client').focus()}
function renderDetail(){const s=state(selected);$('detailClient').textContent=selected.client;$('detailMachine').textContent=selected.machineId;$('detailToken').value=selected.token;$('detailSummary').innerHTML=`<span class="badge ${s}">${labels[s]}</span><span>${remaining(selected)}${selected.permanent?'':` · ${date(selected.expiresAt)} UTC`}</span>`}
function showDetail(row){selected=row;renderDetail();$('detailDialog').showModal()}
async function copyToken(){try{await navigator.clipboard.writeText(selected.token);notify('Token copiado. Listo para entregar al cliente.')}catch(err){$('detailToken').focus();$('detailToken').select();notify('Seleccionado. Presiona Ctrl+C para copiar el token.')}}
$('today').textContent=new Intl.DateTimeFormat('es-BO',{weekday:'long',day:'numeric',month:'long'}).format(new Date());
$('newLicense').onclick=openCreate;$('emptyNew').onclick=openCreate;$('refresh').onclick=load;$('search').oninput=render;
$('navAll').onclick=()=>{filter='all';$('search').value='';render()};
for(const button of document.querySelectorAll('[data-filter]'))button.onclick=()=>{filter=button.dataset.filter;render()};
for(const button of document.querySelectorAll('[data-close]'))button.onclick=()=>$(button.dataset.close).close();
for(const input of document.querySelectorAll('[name=term]'))input.onchange=()=>{const permanent=document.querySelector('[name=term]:checked').value==='permanent';$('dateLabel').hidden=permanent;$('expires').required=!permanent};
$('rows').onclick=e=>{const button=e.target.closest('[data-detail]');if(button){const row=licenses.find(l=>l.id===Number(button.dataset.detail));if(row)showDetail(row)}};
$('createForm').onsubmit=async e=>{e.preventDefault();$('formError').hidden=true;const machineId=$('machine').value.trim();if(!/^[0-9a-f]{64}$/i.test(machineId)){$('formError').textContent='El Machine ID debe contener 64 caracteres hexadecimales.';$('formError').hidden=false;return}
 $('submit').disabled=true;$('submit').textContent='Generando…';try{const row=await api().CreateLicense({client:$('client').value.trim(),machineId,expires:$('expires').value,permanent:document.querySelector('[name=term]:checked').value==='permanent'});$('createDialog').close();await load();showDetail(row);notify('Licencia emitida y guardada correctamente.')}catch(err){$('formError').textContent=String(err?.message||err);$('formError').hidden=false}finally{$('submit').disabled=false;$('submit').textContent='Generar licencia'}};
$('copyDetail').onclick=copyToken;
$('download').onclick=async()=>{if(!selected)return;try{if(await api().ExportLicense(selected.id))notify('Token guardado correctamente.')}catch(err){fail(err)}};
setInterval(()=>{if(licenses.length)render()},30000);load();