import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import type {
  Crossbow,
  SensorData,
  Alert,
  RLStatus,
  RLResult,
  RLTrainingRecord,
  StartSimulationRequest,
  DataQueryRequest,
  APIResponse,
  ArrowTrajectory,
  AlertThresholds
} from '../types'

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

apiClient.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

const handleResponse = <T>(response: AxiosResponse<APIResponse<T>>): T => {
  if (response.data.success) {
    return response.data.data as T
  }
  throw new Error(response.data.message || '请求失败')
}

export const crossbowApi = {
  getList: (): Promise<Crossbow[]> =>
    apiClient.get<APIResponse<Crossbow[]>>('/crossbows').then(handleResponse),

  getById: (id: string): Promise<Crossbow> =>
    apiClient.get<APIResponse<Crossbow>>(`/crossbows/${id}`).then(handleResponse),

  create: (data: Omit<Crossbow, 'id' | 'createdAt' | 'updatedAt'>): Promise<Crossbow> =>
    apiClient.post<APIResponse<Crossbow>>('/crossbows', data).then(handleResponse),

  update: (id: string, data: Partial<Crossbow>): Promise<Crossbow> =>
    apiClient.put<APIResponse<Crossbow>>(`/crossbows/${id}`, data).then(handleResponse),

  delete: (id: string): Promise<void> =>
    apiClient.delete<APIResponse<void>>(`/crossbows/${id}`).then(handleResponse)
}

export const simulationApi = {
  start: (crossbowId: string, request: StartSimulationRequest): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/start`, request).then(handleResponse),

  pause: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/pause`).then(handleResponse),

  resume: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/resume`).then(handleResponse),

  stop: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/stop`).then(handleResponse),

  reset: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/reset`).then(handleResponse),

  setSpeed: (crossbowId: string, speed: number): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/simulation/${crossbowId}/speed`, { speed }).then(handleResponse)
}

export const sensorApi = {
  getLatest: (crossbowId: string): Promise<SensorData> =>
    apiClient.get<APIResponse<SensorData>>(`/sensors/${crossbowId}/latest`).then(handleResponse),

  getHistory: (request: DataQueryRequest): Promise<SensorData[]> =>
    apiClient.post<APIResponse<SensorData[]>>('/sensors/history', request).then(handleResponse),

  getRealtime: (crossbowId: string, limit: number = 100): Promise<SensorData[]> =>
    apiClient.get<APIResponse<SensorData[]>>(`/sensors/${crossbowId}/realtime?limit=${limit}`).then(handleResponse)
}

export const alertApi = {
  getList: (crossbowId?: string, acknowledged?: boolean, limit: number = 100): Promise<Alert[]> => {
    const params = new URLSearchParams()
    if (crossbowId) params.append('crossbowId', crossbowId)
    if (acknowledged !== undefined) params.append('acknowledged', String(acknowledged))
    params.append('limit', String(limit))
    return apiClient.get<APIResponse<Alert[]>>(`/alerts?${params.toString()}`).then(handleResponse)
  },

  acknowledge: (id: string): Promise<Alert> =>
    apiClient.post<APIResponse<Alert>>(`/alerts/${id}/acknowledge`).then(handleResponse),

  getThresholds: (crossbowId: string): Promise<AlertThresholds> =>
    apiClient.get<APIResponse<AlertThresholds>>(`/alerts/thresholds/${crossbowId}`).then(handleResponse),

  updateThresholds: (crossbowId: string, thresholds: Partial<AlertThresholds>): Promise<AlertThresholds> =>
    apiClient.put<APIResponse<AlertThresholds>>(`/alerts/thresholds/${crossbowId}`, thresholds).then(handleResponse)
}

export const rlApi = {
  getStatus: (crossbowId: string): Promise<RLStatus> =>
    apiClient.get<APIResponse<RLStatus>>(`/rl/${crossbowId}/status`).then(handleResponse),

  startTraining: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/rl/${crossbowId}/start`).then(handleResponse),

  stopTraining: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/rl/${crossbowId}/stop`).then(handleResponse),

  getResult: (crossbowId: string): Promise<RLResult> =>
    apiClient.get<APIResponse<RLResult>>(`/rl/${crossbowId}/result`).then(handleResponse),

  getTrainingHistory: (crossbowId: string, limit: number = 100): Promise<RLTrainingRecord[]> =>
    apiClient.get<APIResponse<RLTrainingRecord[]>>(`/rl/${crossbowId}/history?limit=${limit}`).then(handleResponse),

  applyPolicy: (crossbowId: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/rl/${crossbowId}/apply`).then(handleResponse)
}

export const trajectoryApi = {
  getLatest: (crossbowId: string, limit: number = 10): Promise<ArrowTrajectory[]> =>
    apiClient.get<APIResponse<ArrowTrajectory[]>>(`/trajectories/${crossbowId}/latest?limit=${limit}`).then(handleResponse),

  getById: (id: string): Promise<ArrowTrajectory> =>
    apiClient.get<APIResponse<ArrowTrajectory>>(`/trajectories/${id}`).then(handleResponse)
}

export interface CrossbowVariant {
  code: string
  name: string
  dynasty: string
  description: string
  armLength: number
  stringTension: number
  magazineCapacity: number
  maxRange: number
  effectiveRange: number
  theoreticalFireRate: number
  reloadTime: number
  accuracyScore: number
  sustainabilityScore: number
}

export interface VariantCompareResult {
  variants: CrossbowVariant[]
  parameterComparison: {
    parameter: string
    unit: string
    values: Record<string, number>
    bestCode: string
  }[]
  radarData: {
    dimensions: string[]
    series: { code: string; name: string; values: number[] }[]
  }
  advantages: {
    category: string
    code: string
    name: string
    value: number
    unit: string
  }[]
}

export interface ModernFirearm {
  code: string
  name: string
  type: string
  era: number
  theoreticalFireRate: number
  actualFireRate: number
  magazineCapacity: number
  caliber: string
  effectiveRange: number
  muzzleVelocity: number
}

export interface EraCompareResult {
  ancient: CrossbowVariant[]
  modern: ModernFirearm[]
  stats: {
    ancientAvgFireRate: number
    modernAvgFireRate: number
    fireRateRatio: number
    rangeRatio: number
  }
  comparisonTable: {
    name: string
    type: string
    era: number
    theoreticalFireRate: number
    actualFireRate: number
    magazineCapacity: number
    caliber: string
    effectiveRange: number
    muzzleVelocity: number
    relativeRatio: number
  }[]
  timeline: {
    year: number
    name: string
    fireRate: number
    type: 'ancient' | 'modern'
  }[]
}

export interface ReliabilityResult {
  totalShots: number
  jamCount: number
  jamRate: number
  mtbf: number
  mtbfHours: number
  ciLower: number
  ciUpper: number
  reliabilityCurve: { n: number; r: number }[]
  cumulativeJams: { n: number; count: number }[]
  failureModes: { mode: string; name: string; count: number; percentage: number }[]
  fmea: {
    mode: string
    name: string
    severity: number
    occurrence: number
    detection: number
    rpn: number
    suggestion: string
  }[]
}

export interface VirtualShootStartResponse {
  sessionID: string
  variantCode: string
  magazineCapacity: number
  currentAmmo: number
}

export interface VirtualShootResponse {
  shotFired: boolean
  jammed: boolean
  recovered: boolean
  recoverTime: number
  newState: {
    currentAmmo: number
    totalShots: number
    jams: number
    stringFatigue: number
    reloading: boolean
    reloadProgress: number
    cooling: boolean
    coolTimeRemaining: number
  }
  hit?: {
    ring: number
    offsetX: number
    offsetY: number
  }
}

export interface VirtualShootStatus {
  sessionID: string
  variantCode: string
  magazineCapacity: number
  currentAmmo: number
  totalShots: number
  jams: number
  stringFatigue: number
  reloading: boolean
  reloadProgress: number
  cooling: boolean
  coolTimeRemaining: number
  recentShots: number[]
  hitHistory: { ring: number; offsetX: number; offsetY: number }[]
}

export const variantApi = {
  getVariants: (): Promise<CrossbowVariant[]> =>
    apiClient.get<APIResponse<CrossbowVariant[]>>('/variants').then(handleResponse),

  getVariant: (code: string): Promise<CrossbowVariant> =>
    apiClient.get<APIResponse<CrossbowVariant>>(`/variants/${code}`).then(handleResponse),

  compareVariants: (variantCodes: string[]): Promise<VariantCompareResult> =>
    apiClient.post<APIResponse<VariantCompareResult>>('/variants/compare', {
      variantCodes,
      compareMetrics: 'all'
    }).then(handleResponse),

  analyzeReliability: (code: string, params: { shots: number; simTimeSec: number }): Promise<ReliabilityResult> =>
    apiClient.post<APIResponse<ReliabilityResult>>(`/variants/reliability/${code}`, params).then(handleResponse)
}

export const firearmApi = {
  getModernFirearms: (): Promise<ModernFirearm[]> =>
    apiClient.get<APIResponse<ModernFirearm[]>>('/firearms').then(handleResponse),

  compareEra: (params: { ancientVariants: string[]; modernFirearms: string[] }): Promise<EraCompareResult> =>
    apiClient.post<APIResponse<EraCompareResult>>('/firearms/compare-era', params).then(handleResponse)
}

export const virtualShootApi = {
  start: (variantCode: string): Promise<VirtualShootStartResponse> =>
    apiClient.post<APIResponse<VirtualShootStartResponse>>('/virtual/start', { variantCode }).then(handleResponse),

  shoot: (sessionID: string): Promise<VirtualShootResponse> =>
    apiClient.post<APIResponse<VirtualShootResponse>>('/virtual/shoot', {
      sessionID,
      mode: 'single',
      burstCount: 1
    }).then(handleResponse),

  getStatus: (id: string): Promise<VirtualShootStatus> =>
    apiClient.get<APIResponse<VirtualShootStatus>>(`/virtual/${id}`).then(handleResponse),

  reset: (id: string): Promise<void> =>
    apiClient.post<APIResponse<void>>(`/virtual/${id}/reset`).then(handleResponse)
}

export default apiClient
