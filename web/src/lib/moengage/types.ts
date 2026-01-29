export interface MoEngageConfig {
  app_id: string;
  swPath?: string;
  debug?: boolean;
  enable_logging?: boolean;
  utm?: {
    utm_source?: string;
    utm_medium?: string;
    utm_campaign?: string;
  };
}

export interface MoEngageUser {
  name?: string;
  email?: string;
  mobile?: string;
  user_attribute_1?: string;
  user_attribute_2?: string;
  user_attribute_3?: string;
  user_attribute_4?: string;
  user_attribute_5?: string;
}

declare global {
  interface Window {
    Moengage?: MoEngageAPI;
  }
}

export interface MoEngageAPI {
  push: (command: unknown[]) => void;
  track_event?: (eventName: string, eventAttributes?: Record<string, unknown>) => void;
  add_user_attribute?: (attributes: MoEngageUser) => void;
  logout?: () => void;
}
