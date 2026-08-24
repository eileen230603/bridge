/// <reference types="vite/client" />
interface Window {
  go?: {
    main?: {
      App?: {
        SearchStudies(from: string, to: string): Promise<Study[]>;
        CreateDiscJob(uid: string): Promise<DiscJob>;
        ListJobs(): Promise<DiscJob[]>;
        TestPacsConnection(): Promise<void>;
        GetSystemStatus(): Promise<{
          studyServer: string;
          studyApiConfigured: boolean;
          pacs: string;
          pacsConfigured: boolean;
          tdBridge: string;
          tdBridgeConfigured: boolean;
        }>;
      };
    };
  };
}
