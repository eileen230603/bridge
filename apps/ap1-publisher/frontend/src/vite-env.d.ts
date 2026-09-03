/// <reference types="vite/client" />
interface ServerConfig { protocol: "http" | "https"; host: string; port: number; timeoutSeconds: number }
interface ConnectionTestResult { status: string; message: string }
<<<<<<< HEAD
=======
interface DiscLabelConfig { hospitalName: string; logoPath: string }
>>>>>>> origin/eileen
interface Window {
  go?: {
    main?: {
      App?: {
        SearchStudies(from: string, to: string): Promise<Study[]>;
        CancelSearch(): Promise<void>;
        CreateDiscJob(uid: string): Promise<DiscJob>;
        ListJobs(): Promise<DiscJob[]>;
        GetSystemStatus(): Promise<{
          studyServer: string;
          studyServerAddress: string;
          studyApiConfigured: boolean;
          tdBridge: string;
          tdBridgeConfigured: boolean;
        }>;
        GetServerConfig(): Promise<ServerConfig>;
        SaveServerConfig(config: ServerConfig): Promise<void>;
        TestServerConnection(config: ServerConfig): Promise<ConnectionTestResult>;
<<<<<<< HEAD
=======
        GetDiscLabelConfig(): Promise<DiscLabelConfig>;
        SaveDiscLabelConfig(config: DiscLabelConfig): Promise<void>;
        SelectDiscLabelLogo(): Promise<string>;
        GetDiscLabelPreview(config: DiscLabelConfig): Promise<string>;
>>>>>>> origin/eileen
      };
    };
  };
}
