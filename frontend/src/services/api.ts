import axios, {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig,
} from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private client: AxiosInstance;
  private tokenGetter: (() => Promise<string>) | null = null;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.client.interceptors.request.use(
      async (config: InternalAxiosRequestConfig) => {
        if (this.tokenGetter) {
          const token = await this.tokenGetter();
          if (token) {
            config.headers.Authorization = `Bearer ${token}`;
          }
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Token expired or invalid - trigger re-authentication
          this.handleUnauthorized();
        }
        return Promise.reject(error);
      }
    );
  }

  public setTokenGetter(getter: () => Promise<string>): void {
    this.tokenGetter = getter;
  }

  public setUnauthorizedHandler(handler: () => void): void {
    this._unauthorizedHandler = handler;
  }

  private _unauthorizedHandler: (() => void) | null = null;

  private handleUnauthorized(): void {
    if (this._unauthorizedHandler) {
      this._unauthorizedHandler();
    }
  }

  public async get<T = any>(
    url: string,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse<T>> {
    return this.client.get<T>(url, config);
  }

  public async post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse<T>> {
    return this.client.post<T>(url, data, config);
  }

  public async put<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse<T>> {
    return this.client.put<T>(url, data, config);
  }

  public async delete<T = any>(
    url: string,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse<T>> {
    return this.client.delete<T>(url, config);
  }

  public async patch<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse<T>> {
    return this.client.patch<T>(url, data, config);
  }

  public getClient(): AxiosInstance {
    return this.client;
  }
}

export const apiClient = new ApiClient();
