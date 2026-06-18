package dynamics_engine

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// DynamicsTask 每个射击请求的物理计算任务
type DynamicsTask struct {
	TaskID        string
	SessionID     string
	VariantCode   string
	ShotsInBurst  int
	// 输入机构参数
	BowArmLengthM  float64 // L  (m)
	StringTensionN float64 // T0 (N)
	ArrowMassG     float64 // m_arrow (g)
	TriggerForceN  float64 // 扳机峰值力
	BowMassKG      float64 // 弩臂等效质量
	// 环境
	AirDensity    float64 // kg/m³, 默认1.225
	Gravity       float64 // m/s², 默认9.81
	TempCelsius   float64
	HumidityPct   float64
	WindMPS       float64 // 横向风速
	// 计算结果（填充后用Done返回）
	Result *DynamicsResult
	Done   chan struct{}
}

// DynamicsResult 多刚体动力学输出
type DynamicsResult struct {
	// 箭体运动学
	MuzzleVelocityMPS   float64   // 出口初速 (m/s)
	KineticEnergyJ      float64   // 出口动能 (J)
	MuzzleImpulseNs     float64   // 冲量 (N·s)
	MaxRangeM           float64   // 理论最大射程
	TimeOfFlightSec     float64   // 最大射程飞行时间
	ArrowSpinHz         float64   // 箭支气动自转估计
	// 弩臂振动（自由-自由梁）
	BowFreqHz           float64   // 一阶弯振主频
	BowDampingRatio     float64   // 阻尼比
	BowAmpMM            float64   // 振幅(mm)
	VibrationDecayMs    float64   // 衰减到1%所需时间
	// 弦运动
	StringVelocityMPS   float64   // 弦回程最大速度
	StringTensionPeakN  float64   // 释放后峰值张力
	StringTravelM       float64   // 弦总行程
	// 扳机力学
	ReleaseLatencyMs    float64   // 触发到释放延迟
	TriggerWorkJ        float64   // 扣压做功
	TriggerImpulseNs    float64   // 扣压冲量
	// 能量转化
	EfficiencyPct       float64   // 弓弦→箭能量效率
	HeatLossW           float64   // 摩擦热损失功率
	SoundLevelDb        float64   // 发射噪声估计
	// 多体状态序列（可选20帧动画关键帧）
	KeyFrames           []RigidBodyFrame
	// 误差
	Error               string    `json:",omitempty"`
	ComputeMs           int64     `json:"-"`
}

// RigidBodyFrame 多刚体快照（每帧一个时间戳，用于前端Canvas/Three.js动画）
type RigidBodyFrame struct {
	TimeSec     float64 `json:"t"`
	// 弓臂
	BowAngleDeg float64 `json:"bowA"` // 弓臂张开角度
	BowTipX     float64 `json:"bowTipX"`
	BowTipY     float64 `json:"bowTipY"`
	// 弦
	StringMidX  float64 `json:"strX"`
	StringMidY  float64 `json:"strY"`
	StringTension float64 `json:"strT"`
	// 箭
	ArrowX      float64 `json:"arrX"`
	ArrowY      float64 `json:"arrY"`
	ArrowVX     float64 `json:"arrVx"`
	ArrowVY     float64 `json:"arrVy"`
}

type EngineStats struct {
	TasksTotal    int64
	TasksDone     int64
	TasksRejected int64
	AvgLatencyMs  float64
	Workers       int
}

// Engine 多刚体动力学引擎（常驻独立goroutine worker pool）
type Engine struct {
	tasks    chan *DynamicsTask
	wg       sync.WaitGroup
	workers  int
	stop     chan struct{}
	stats    EngineStats
	running  bool
	mu       sync.RWMutex
}

var (
	defaultOnce  sync.Once
	defaultEng   *Engine
)

func newEngine(workers, queue int) *Engine {
	if workers <= 0 { workers = 2 }
	if queue <= 0 { queue = 256 }
	e := &Engine{
		tasks:   make(chan *DynamicsTask, queue),
		stop:    make(chan struct{}),
		workers: workers,
	}
	e.running = true
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
	return e
}

// DefaultEngine 单例
func DefaultEngine() *Engine {
	defaultOnce.Do(func() {
		defaultEng = newEngine(2, 512)
	})
	return defaultEng
}

// NewEngine 构造自定义大小引擎
func NewEngine(workers, queue int) *Engine {
	return newEngine(workers, queue)
}

func (e *Engine) Submit(t *DynamicsTask) error {
	e.mu.RLock()
	running := e.running
	e.mu.RUnlock()
	if !running {
		// 引擎已停止：同步回退计算
		computeDynamics(t)
		return nil
	}
	atomic.AddInt64(&e.stats.TasksTotal, 1)
	select {
	case e.tasks <- t:
		return nil
	default:
		atomic.AddInt64(&e.stats.TasksRejected, 1)
		// 队列满：同步回退，避免阻塞
		computeDynamics(t)
		return nil
	}
}

// SubmitSync 同步计算
func (e *Engine) SubmitSync(t *DynamicsTask) *DynamicsResult {
	t.Done = make(chan struct{}, 1)
	_ = e.Submit(t)
	select {
	case <-t.Done:
	case <-time.After(1500 * time.Millisecond):
		if t.Result == nil {
			computeDynamics(t)
			if t.Done != nil {
				select { case <-t.Done: default: }
			}
		}
	}
	return t.Result
}

func (e *Engine) Stats() EngineStats {
	s := EngineStats{
		TasksTotal:    atomic.LoadInt64(&e.stats.TasksTotal),
		TasksDone:     atomic.LoadInt64(&e.stats.TasksDone),
		TasksRejected: atomic.LoadInt64(&e.stats.TasksRejected),
		Workers:       e.workers,
	}
	return s
}

func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running { e.mu.Unlock(); return }
	e.running = false
	close(e.stop)
	e.mu.Unlock()
	e.wg.Wait()
}

func (e *Engine) worker(id int) {
	defer e.wg.Done()
	for {
		select {
		case <-e.stop:
			return
		case t, ok := <-e.tasks:
			if !ok { return }
			start := time.Now()
			computeDynamics(t)
			atomic.AddInt64(&e.stats.TasksDone, 1)
			elapsed := time.Since(start).Milliseconds()
			// 滑动平均
			e.mu.Lock()
			if e.stats.AvgLatencyMs == 0 {
				e.stats.AvgLatencyMs = float64(elapsed)
			} else {
				e.stats.AvgLatencyMs = 0.9*e.stats.AvgLatencyMs + 0.1*float64(elapsed)
			}
			e.mu.Unlock()
			if t.Result != nil { t.Result.ComputeMs = elapsed }
			if t.Done != nil {
				select { case t.Done <- struct{}{}: default: close(t.Done) }
			}
		}
	}
}

// ============================================================
// 核心多刚体物理计算
// ============================================================

func computeDynamics(t *DynamicsTask) {
	start := time.Now()
	if t == nil { return }
	if t.BowArmLengthM <= 0 { t.BowArmLengthM = 0.5 }
	if t.StringTensionN <= 1 { t.StringTensionN = 900 }
	if t.ArrowMassG <= 0.1 { t.ArrowMassG = 50 }
	if t.AirDensity <= 0 { t.AirDensity = 1.225 }
	if t.Gravity <= 0 { t.Gravity = 9.81 }
	if t.BowMassKG <= 0 { t.BowMassKG = 1.5 }

	mArrowKG := t.ArrowMassG * 0.001

	// 1. 弹性梁弦-箭耦合：近似能量守恒
	// 弓弦变形能 U = 0.5 * T0 * draw; draw = 0.7 * BowArmLengthM
	draw := 0.7 * t.BowArmLengthM
	energySpring := 0.5 * t.StringTensionN * draw
	// 经验效率：28~42%取决于弓/箭质量比
	massRatio := mArrowKG / math.Max(0.001, t.BowMassKG)
	eff := 0.28 + 0.14*(1.0-math.Exp(-massRatio*20))
	if eff < 0.25 { eff = 0.25 }
	if eff > 0.48 { eff = 0.48 }
	energyArrow := energySpring * eff
	vMuzzle := math.Sqrt(2 * energyArrow / mArrowKG)
	if vMuzzle > 250 { vMuzzle = 250 }
	if vMuzzle < 20 { vMuzzle = 20 }

	// 2. 外弹道：真空近似 + 45度仰角射程
	maxRange := vMuzzle * vMuzzle / t.Gravity
	tof := 2 * vMuzzle * math.Sin(math.Pi/4) / t.Gravity

	// 3. 弩臂一阶弯振：伯努利-欧拉梁 f = (1.875² / 2π) * √(EI/ρAL⁴)
	// 简化: 给定经验常数，L越小→频率越高，张力越大→频率越高
	kSpring := t.StringTensionN / math.Max(0.001, t.BowArmLengthM)
	fBow := (1 / (2 * math.Pi)) * math.Sqrt(kSpring/math.Max(0.001, t.BowMassKG)) * 3.5
	if fBow < 4 { fBow = 4 }
	if fBow > 50 { fBow = 50 }
	damp := 0.06 + 0.04*math.Tanh(t.StringTensionN/3000)
	decayMs := 3.0 / (2 * math.Pi * fBow * damp) * 1000
	ampMM := 2.0 + 6.0*(t.StringTensionN/4000)*(vMuzzle/120)

	// 4. 弦运动学
	vString := vMuzzle * (0.75 + 0.15*massRatio)
	stringTravel := draw * 1.08
	tensionPeak := t.StringTensionN * (1.15 + 0.2*math.Tanh(massRatio*15))

	// 5. 扳机延迟：力越大→释放越快（机械放大杠杆）
	latencyMs := 80.0
	if t.TriggerForceN > 1 { latencyMs = 12 + 80/math.Sqrt(t.TriggerForceN/10) }
	if latencyMs < 5 { latencyMs = 5 }
	if latencyMs > 250 { latencyMs = 250 }

	// 6. 扳机功/冲量估算
	triggerTravelM := 0.005 + 0.003*math.Tanh(t.TriggerForceN/200)
	workJ := t.TriggerForceN * triggerTravelM * 0.5
	impNs := t.TriggerForceN * (latencyMs / 1000.0) * 0.4

	// 7. 损失与噪声
	heatW := (energySpring - energyArrow) / math.Max(0.005, (1/fBow))
	soundDb := 65.0 + 25*math.Log10(math.Max(1, vMuzzle/50))

	// 8. 箭支自旋：由羽毛角度诱导（2度预扭→约50Hz自旋）
	spinHz := 45 + 5*vMuzzle/80
	if spinHz < 20 { spinHz = 20 }

	// 9. 关键帧（弦释放→箭出口 20帧）
	nF := 20
	frames := make([]RigidBodyFrame, 0, nF)
	durMs := 0.012 + 0.008*t.BowArmLengthM/0.6
	bowInitialAngle := 25.0 + 30*t.BowArmLengthM/0.9
	for i := 0; i < nF; i++ {
		f := float64(i) / float64(nF-1)
		tSec := f * durMs
		// 弦位移：余弦加速
		normT := f
		x := 1 - math.Cos(normT*math.Pi/2) // 0..1
		bowAngle := bowInitialAngle * (1 - 0.75*x)
		bowTipX := t.BowArmLengthM * math.Sin(bowAngle*math.Pi/180) * 1000 // mm
		bowTipY := t.BowArmLengthM * math.Cos(bowAngle*math.Pi/180) * 1000
		strX := 0.0
		strY := (0.30 + 0.70*x) * stringTravel * 1000
		strT := t.StringTensionN * (1 - 0.5*x)
		arrX := strY
		arrY := 0.0
		arrV := vMuzzle * math.Sqrt(x)
		frames = append(frames, RigidBodyFrame{
			TimeSec:     tSec,
			BowAngleDeg: bowAngle,
			BowTipX:     bowTipX, BowTipY: bowTipY,
			StringMidX: strX, StringMidY: strY, StringTension: strT,
			ArrowX: arrX, ArrowY: arrY, ArrowVX: arrV, ArrowVY: 0.0,
		})
	}

	r := &DynamicsResult{
		MuzzleVelocityMPS: vMuzzle,
		KineticEnergyJ:    energyArrow,
		MuzzleImpulseNs:   mArrowKG * vMuzzle * math.Max(1, float64(t.ShotsInBurst)),
		MaxRangeM:         maxRange,
		TimeOfFlightSec:   tof,
		ArrowSpinHz:       spinHz,
		BowFreqHz:         fBow,
		BowDampingRatio:   damp,
		BowAmpMM:          ampMM,
		VibrationDecayMs:  decayMs,
		StringVelocityMPS: vString,
		StringTensionPeakN: tensionPeak,
		StringTravelM:     stringTravel,
		ReleaseLatencyMs:  latencyMs,
		TriggerWorkJ:      workJ,
		TriggerImpulseNs:  impNs,
		EfficiencyPct:     eff * 100,
		HeatLossW:         heatW,
		SoundLevelDb:      soundDb,
		KeyFrames:         frames,
	}
	r.ComputeMs = time.Since(start).Milliseconds()
	t.Result = r
}

// ================= 便捷工具函数 ====================

type VariantProfile struct {
	VariantCode  string
	BowArmLengthM float64
	StringTensionN float64
	ArrowMassG    float64
	BowMassKG     float64
}

func LookupVariantProfile(code string) VariantProfile {
	switch code {
	case "zhuge":
		return VariantProfile{code, 0.45, 950, 50, 1.2}
	case "san-gong":
		return VariantProfile{code, 0.90, 3500, 350, 8.5}
	case "bi-zhang":
		return VariantProfile{code, 0.60, 1500, 72, 2.4}
	default:
		return VariantProfile{code, 0.5, 900, 60, 1.5}
	}
}

// NewTaskFromVariant 从弩型代码构建任务
func NewTaskFromVariant(taskID, sessionID, variantCode string, shotsInBurst int, triggerForceN float64) *DynamicsTask {
	p := LookupVariantProfile(variantCode)
	return &DynamicsTask{
		TaskID: taskID, SessionID: sessionID, VariantCode: variantCode,
		ShotsInBurst: shotsInBurst,
		BowArmLengthM: p.BowArmLengthM, StringTensionN: p.StringTensionN,
		ArrowMassG: p.ArrowMassG, BowMassKG: p.BowMassKG,
		TriggerForceN: triggerForceN,
		AirDensity: 1.225, Gravity: 9.81,
		TempCelsius: 20, HumidityPct: 50, WindMPS: 0,
	}
}
