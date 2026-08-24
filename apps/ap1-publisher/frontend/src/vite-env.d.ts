/// <reference types="vite/client" />
interface Window {
  go?: {
    main?: {
      App?: {
        SearchStudies(from: string, to: string): Promise<Study[]>;
        CreateDiscJob(uid: string): Promise<DiscJob>;
        ListJobs(): Promise<DiscJob[]>;
        GetSystemStatus(): Promise<{
          studyServer: string;
          studyApiConfigured: boolean;
          tdBridge: string;
          tdBridgeConfigured: boolean;
        }>;
      };
    };
  };
}
