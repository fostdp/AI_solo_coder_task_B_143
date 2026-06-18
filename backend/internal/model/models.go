package model

import (
	"encoding/json"
	"time"
)

type CrossbowConfig struct {
	BowArmLength        float64 `json:"bowArmLength"`
	BowArmStiffness     float64 `json:"bowArmStiffness"`
	StringLength        float64 `json:"stringLength"`
	StringTension       float64 `json:"stringTension"`
	StringFatigueLimit  float64 `json:"stringFatigueLimit"`
	ArrowMass           float64 `json:"arrowMass"`
	MagazineCapacity    int     `json:"magazineCapacity"`
	CamRadius           float64 `json:"camRadius"`
	CamLift             float64 `json:"camLift"`
	FrictionCoefficient float64 `json:"frictionCoefficient"`
	Gravity             float64 `json:"gravity"`
}

type Crossbow struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Config      CrossbowConfig  `json:"config"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type SensorData struct {
	CrossbowID        string    `json:"crossbowId"`
	Timestamp         time.Time `json:"timestamp"`
	StringTension     float64   `json:"stringTension"`
	BowArmDeformation float64   `json:"bowArmDeformation"`
	MagazinePosition  float64   `json:"magazinePosition"`
	FireRate          float64   `json:"fireRate"`
	ArrowVelocity     float64   `json:"arrowVelocity"`
	CamAngle          float64   `json:"camAngle"`
	StringFatigue     float64   `json:"stringFatigue"`
	Temperature       float64   `json:"temperature"`
}

type DynamicsState struct {
	Timestamp           time.Time       `json:"timestamp"`
	BowArmAngle         float64         `json:"bowArmAngle"`
	BowArmAngularVel    float64         `json:"bowArmAngularVel"`
	BowArmAngularAcc    float64         `json:"bowArmAngularAcc"`
	StringDisplacement  float64         `json:"stringDisplacement"`
	StringVelocity      float64         `json:"stringVelocity"`
	CamPosition         float64         `json:"camPosition"`
	PawlEngaged         bool            `json:"pawlEngaged"`
	LoadingComplete     bool            `json:"loadingComplete"`
	ArrowLoaded         bool            `json:"arrowLoaded"`
	Forces              json.RawMessage `json:"forces,omitempty"`
}

type TrajectoryPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	T float64 `json:"t"`
}

type ArrowTrajectory struct {
	ID              string            `json:"id"`
	CrossbowID      string            `json:"crossbowId"`
	FireTime        time.Time         `json:"fireTime"`
	Positions       []TrajectoryPoint `json:"positions"`
	InitialVelocity float64           `json:"initialVelocity"`
	FlightTime      float64           `json:"flightTime"`
	ImpactPoint     TrajectoryPoint   `json:"impactPoint"`
	CreatedAt       time.Time         `json:"createdAt"`
}

type Alert struct {
	ID             string    `json:"id"`
	CrossbowID     string    `json:"crossbowId"`
	Type           string    `json:"type"`
	Level          string    `json:"level"`
	Message        string    `json:"message"`
	Value          float64   `json:"value"`
	Threshold      float64   `json:"threshold"`
	CreatedAt      time.Time `json:"createdAt"`
	Acknowledged   bool      `json:"acknowledged"`
	AcknowledgedAt time.Time `json:"acknowledgedAt,omitempty"`
}

type AlertThresholds struct {
	ID                  string    `json:"id"`
	CrossbowID          string    `json:"crossbowId"`
	StringTensionMax    float64   `json:"stringTensionMax"`
	StringFatigueWarning float64  `json:"stringFatigueWarning"`
	FireRateMin         float64   `json:"fireRateMin"`
	DeformationMax      float64   `json:"deformationMax"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// FatigueState 疲劳状态（与simulation.FatigueState兼容）
type FatigueState struct {
	StringFatigue       float64     `json:"stringFatigue"`
	TotalDamage         float64     `json:"totalDamage"`
	MaxStress           float64     `json:"maxStress"`
	Cycles              float64     `json:"cycles"`
	TotalDeltaL         float64     `json:"totalDeltaL"`
	CurrentLifeFraction float64     `json:"currentLifeFraction"`
	TensionHistory      []float64   `json:"tensionHistory,omitempty"`
	LastUpdated         time.Time   `json:"lastUpdated"`
}

type RLStatus struct {
	IsTraining        bool      `json:"isTraining"`
	Episode           int       `json:"episode"`
	TotalReward       float64   `json:"totalReward"`
	AverageReward     float64   `json:"averageReward"`
	Epsilon           float64   `json:"epsilon"`
	CurrentPolicy     []float64 `json:"currentPolicy"`
	TrainingStartTime time.Time `json:"trainingStartTime,omitempty"`
	BestReward        float64   `json:"bestReward"`
}

type RLResult struct {
	ID                         string    `json:"id"`
	CrossbowID                 string    `json:"crossbowId"`
	OptimizedFireRate          float64   `json:"optimizedFireRate"`
	OptimizedLoadingInterval   float64   `json:"optimizedLoadingInterval"`
	FatigueReduction           float64   `json:"fatigueReduction"`
	EfficiencyImprovement      float64   `json:"efficiencyImprovement"`
	SustainedFireDuration      float64   `json:"sustainedFireDuration"`
	TrainingEpisodes           int       `json:"trainingEpisodes"`
	ConvergenceReward          float64   `json:"convergenceReward"`
	FinalPolicy                []float64 `json:"finalPolicy"`
	CreatedAt                  time.Time `json:"createdAt"`
}

type RLTrainingRecord struct {
	ID            string    `json:"id"`
	CrossbowID    string    `json:"crossbowId"`
	Episode       int       `json:"episode"`
	TotalReward   float64   `json:"totalReward"`
	AverageReward float64   `json:"averageReward"`
	Epsilon       float64   `json:"epsilon"`
	Policy        []float64 `json:"policy"`
	CreatedAt     time.Time `json:"createdAt"`
}

type WSMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

type StartSimulationRequest struct {
	SimulationSpeed float64 `json:"simulationSpeed"`
	EnableRL        bool    `json:"enableRL"`
	Duration        int     `json:"duration"`
}

type DataQueryRequest struct {
	CrossbowID  string   `json:"crossbowId"`
	StartTime   string   `json:"startTime"`
	EndTime     string   `json:"endTime"`
	Metrics     []string `json:"metrics"`
	Aggregation string   `json:"aggregation"`
	Interval    string   `json:"interval"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func DefaultCrossbowConfig() CrossbowConfig {
	return CrossbowConfig{
		BowArmLength:        0.65,
		BowArmStiffness:     5000,
		StringLength:        0.85,
		StringTension:       950,
		StringFatigueLimit:  0.9,
		ArrowMass:           0.05,
		MagazineCapacity:    10,
		CamRadius:           0.04,
		CamLift:             0.12,
		FrictionCoefficient: 0.15,
		Gravity:             9.81,
	}
}

type VariantRecord struct {
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Dynasty         string          `json:"dynasty,omitempty"`
	EraYear         int             `json:"eraYear,omitempty"`
	Description     string          `json:"description,omitempty"`
	DrawWeightN     float64         `json:"drawWeightN,omitempty"`
	MaxRangeM       float64         `json:"maxRangeM,omitempty"`
	EffectiveRangeM float64         `json:"effectiveRangeM,omitempty"`
	IdealFireRate   float64         `json:"idealFireRate,omitempty"`
	MagazineSize    int             `json:"magazineSize,omitempty"`
	ReloadTimeSec   float64         `json:"reloadTimeSec,omitempty"`
	AccuracyScore   float64         `json:"accuracyScore,omitempty"`
	MechanismParams json.RawMessage `json:"mechanismParams,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type FirearmRecord struct {
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	Origin            string    `json:"origin,omitempty"`
	IntroYear         int       `json:"introYear,omitempty"`
	Type              string    `json:"type,omitempty"`
	CyclicRateRPM     int       `json:"cyclicRateRpm,omitempty"`
	EffectiveRPM      int       `json:"effectiveRpm,omitempty"`
	MagazineSize      int       `json:"magazineSize,omitempty"`
	CaliberMM         float64   `json:"caliberMm,omitempty"`
	EffectiveRangeM   int       `json:"effectiveRangeM,omitempty"`
	MuzzleVelocityMPS float64   `json:"muzzleVelocityMps,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

type ReliabilityAnalysisRecord struct {
	ID                     string          `json:"id"`
	CrossbowVariantCode    string          `json:"crossbowVariantCode"`
	SimShots               int             `json:"simShots"`
	SimTimeSec             float64         `json:"simTimeSec"`
	JamProbabilityPerShot  float64         `json:"jamProbabilityPerShot,omitempty"`
	MTBFShots              float64         `json:"mtbfShots,omitempty"`
	MTBFHours              float64         `json:"mtbfHours,omitempty"`
	FailureModeDistribution json.RawMessage `json:"failureModeDistribution,omitempty"`
	ReliabilityCurve       json.RawMessage `json:"reliabilityCurve,omitempty"`
	ConfidenceInterval     json.RawMessage `json:"confidenceInterval,omitempty"`
	CreatedAt              time.Time       `json:"createdAt"`
}

type VirtualShootSessionRecord struct {
	SessionID          string      `json:"sessionId"`
	CrossbowVariantCode string     `json:"crossbowVariantCode"`
	UserID             string      `json:"userId,omitempty"`
	ShotsFired         int         `json:"shotsFired,omitempty"`
	JamCount           int         `json:"jamCount,omitempty"`
	ReloadCount        int         `json:"reloadCount,omitempty"`
	ElapsedSec         float64     `json:"elapsedSec,omitempty"`
	AverageRPM         float64     `json:"averageRpm,omitempty"`
	MaxInstantRPM      float64     `json:"maxInstantRpm,omitempty"`
	FinalStringFatigue float64     `json:"finalStringFatigue,omitempty"`
	StartedAt          time.Time   `json:"startedAt,omitempty"`
	EndedAt            *time.Time  `json:"endedAt,omitempty"`
	ShotTimestamps     []time.Time `json:"shotTimestamps,omitempty"`
}

type MechanismParams struct {
	BowArmLength  float64 `json:"bowArmLength"`
	StringTension float64 `json:"stringTension"`
	MagazineCap   int     `json:"magazineCap"`
	ReloadTimeSec float64 `json:"reloadTimeSec"`
}

type PerformanceMetrics struct {
	DrawWeight      float64 `json:"drawWeight"`
	MaxRange        float64 `json:"maxRange"`
	EffectiveRange  float64 `json:"effectiveRange"`
	IdealFireRate   float64 `json:"idealFireRate"`
	MagazineSize    int     `json:"magazineSize"`
	ReloadTime      float64 `json:"reloadTime"`
	AccuracyScore   float64 `json:"accuracyScore"`
}

type CrossbowVariant struct {
	VariantCode      string            `json:"variantCode"`
	Name             string            `json:"name"`
	Dynasty          string            `json:"dynasty"`
	Description      string            `json:"description"`
	MechanismParams  *MechanismParams  `json:"mechanismParams"`
	Performance      PerformanceMetrics `json:"performance"`
}

type VariantCompareRequest struct {
	VariantCodes   []string `json:"variantCodes"`
	CompareMetrics []string `json:"compareMetrics"`
}

type PerformanceRadar struct {
	Metric string             `json:"metric"`
	Values map[string]float64 `json:"values"`
	Best   string             `json:"best"`
}

type AdvantageMap struct {
	Metric         string   `json:"metric"`
	BestVariant    string   `json:"bestVariant"`
	BestValue      float64  `json:"bestValue"`
	RunnerUp       string   `json:"runnerUp"`
	AdvantageRatio float64  `json:"advantageRatio"`
}

type VariantCompareResponse struct {
	ComparedVariants []CrossbowVariant `json:"comparedVariants"`
	PerformanceRadar []PerformanceRadar `json:"performanceRadar"`
	AdvantageMap     []AdvantageMap    `json:"advantageMap"`
}

func CrossbowPresets() []CrossbowVariant {
	return []CrossbowVariant{
		{
			VariantCode: "zhuge",
			Name:        "诸葛弩",
			Dynasty:     "三国·蜀汉",
			Description: "诸葛亮发明的连发弩，又称元戎弩，采用连杆式箭匣供弹，可连续发射十支箭，是古代速射武器的代表。",
			MechanismParams: &MechanismParams{
				BowArmLength:  0.45,
				StringTension: 950,
				MagazineCap:   10,
				ReloadTimeSec: 8.0,
			},
			Performance: PerformanceMetrics{
				DrawWeight:     950,
				MaxRange:       150,
				EffectiveRange: 80,
				IdealFireRate:  10.5,
				MagazineSize:   10,
				ReloadTime:     8.0,
				AccuracyScore:  0.62,
			},
		},
		{
			VariantCode: "san-gong",
			Name:        "三弓弩",
			Dynasty:     "北宋",
			Description: "床子弩的一种，以三张弓合力发射，威力巨大，又称八牛弩，需绞车上弦，专用于攻城破阵。",
			MechanismParams: &MechanismParams{
				BowArmLength:  0.9,
				StringTension: 3500,
				MagazineCap:   1,
				ReloadTimeSec: 45.0,
			},
			Performance: PerformanceMetrics{
				DrawWeight:     3500,
				MaxRange:       500,
				EffectiveRange: 300,
				IdealFireRate:  1.5,
				MagazineSize:   1,
				ReloadTime:     45.0,
				AccuracyScore:  0.88,
			},
		},
		{
			VariantCode: "bi-zhang",
			Name:        "臂张弩",
			Dynasty:     "战国·秦",
			Description: "以手臂力量上弦的单兵弩，秦弩的典型代表，射法简便，射程远于弓箭，是秦军制式装备。",
			MechanismParams: &MechanismParams{
				BowArmLength:  0.6,
				StringTension: 1500,
				MagazineCap:   1,
				ReloadTimeSec: 15.0,
			},
			Performance: PerformanceMetrics{
				DrawWeight:     1500,
				MaxRange:       250,
				EffectiveRange: 150,
				IdealFireRate:  4.0,
				MagazineSize:   1,
				ReloadTime:     15.0,
				AccuracyScore:  0.78,
			},
		},
	}
}

type ModernFirearm struct {
	Name               string  `json:"name"`
	Origin             string  `json:"origin"`
	Era                string  `json:"era"`
	FirearmType        string  `json:"firearmType"`
	CyclicRateRPM      float64 `json:"cyclicRateRPM"`
	EffectiveRPM       float64 `json:"effectiveRPM"`
	MagazineSize       int     `json:"magazineSize"`
	Caliber            string  `json:"caliber"`
	EffectiveRangeM    float64 `json:"effectiveRangeM"`
	MuzzleVelocityMPS  float64 `json:"muzzleVelocityMPS"`
	Notes              string  `json:"notes"`
}

type EraFirearmCompareRequest struct {
	AncientVariants []string `json:"ancientVariants"`
	ModernFirearms  []string `json:"modernFirearms"`
	CompareMetrics  []string `json:"compareMetrics"`
}

type EraGapEntry struct {
	Metric        string  `json:"metric"`
	AncientValue  float64 `json:"ancientValue"`
	AncientUnit   string  `json:"ancientUnit"`
	ModernValue   float64 `json:"modernValue"`
	ModernUnit    string  `json:"modernUnit"`
	GapRatio      float64 `json:"gapRatio"`
	Remark        string  `json:"remark"`
}

type EraFirearmCompareResponse struct {
	AncientVariants []CrossbowVariant `json:"ancientVariants"`
	ModernFirearms  []ModernFirearm   `json:"modernFirearms"`
	EraGapTable     []EraGapEntry     `json:"eraGapTable"`
}

func ModernFirearmPresets() []ModernFirearm {
	return []ModernFirearm{
		{
			Name:              "M1 Garand",
			Origin:            "美国",
			Era:               "1936",
			FirearmType:       "半自动步枪",
			CyclicRateRPM:     0,
			EffectiveRPM:      45,
			MagazineSize:      8,
			Caliber:           ".30-06 Springfield",
			EffectiveRangeM:   457,
			MuzzleVelocityMPS: 853,
			Notes:             "二战美军制式半自动步枪，8发漏夹供弹，被巴顿将军称为\"史上最伟大的战斗工具\"。",
		},
		{
			Name:              "AK-47",
			Origin:            "苏联",
			Era:               "1947",
			FirearmType:       "突击步枪",
			CyclicRateRPM:     600,
			EffectiveRPM:      100,
			MagazineSize:      30,
			Caliber:           "7.62×39mm",
			EffectiveRangeM:   300,
			MuzzleVelocityMPS: 715,
			Notes:             "卡拉什尼科夫设计的经典突击步枪，全球产量超1亿支，以可靠性和简单易用著称。",
		},
		{
			Name:              "M16A1",
			Origin:            "美国",
			Era:               "1967",
			FirearmType:       "突击步枪",
			CyclicRateRPM:     825,
			EffectiveRPM:      52,
			MagazineSize:      20,
			Caliber:           "5.56×45mm M193",
			EffectiveRangeM:   460,
			MuzzleVelocityMPS: 991,
			Notes:             "越南战争时期美军制式突击步枪，开创小口径高速弹先河，精度优良。",
		},
		{
			Name:              "HK MP5",
			Origin:            "西德",
			Era:               "1966",
			FirearmType:       "冲锋枪",
			CyclicRateRPM:     800,
			EffectiveRPM:      100,
			MagazineSize:      30,
			Caliber:           "9×19mm Parabellum",
			EffectiveRangeM:   100,
			MuzzleVelocityMPS: 400,
			Notes:             "HK公司基于G3步枪研发的冲锋枪，全球反恐特种部队标配，射击精度极高。",
		},
		{
			Name:              "M249 SAW",
			Origin:            "美国(比利时设计)",
			Era:               "1984",
			FirearmType:       "班用自动武器/轻机枪",
			CyclicRateRPM:     875,
			EffectiveRPM:      200,
			MagazineSize:      200,
			Caliber:           "5.56×45mm NATO",
			EffectiveRangeM:   800,
			MuzzleVelocityMPS: 915,
			Notes:             "美军步兵班火力支柱，弹链供弹，可换枪管，持续火力压制能力极强。",
		},
		{
			Name:              "Desert Eagle",
			Origin:            "以色列/美国",
			Era:               "1983",
			FirearmType:       "半自动手枪",
			CyclicRateRPM:     0,
			EffectiveRPM:      30,
			MagazineSize:      7,
			Caliber:           ".50 Action Express",
			EffectiveRangeM:   50,
			MuzzleVelocityMPS: 470,
			Notes:             "IMI与Magnum Research联合设计，采用导气式自动原理，是量产手枪中口径最大的型号之一。",
		},
	}
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Event     string  `json:"event,omitempty"`
}

type VirtualShootSession struct {
	SessionID          string            `json:"sessionId"`
	CrossbowVariantCode string           `json:"crossbowVariantCode"`
	ShotsFired         int               `json:"shotsFired"`
	JamCount           int               `json:"jamCount"`
	ReloadCount        int               `json:"reloadCount"`
	ElapsedSec         float64           `json:"elapsedSec"`
	InstantaneousRPM   float64           `json:"instantaneousRPM"`
	AverageRPM         float64           `json:"averageRPM"`
	CurrentAmmo        int               `json:"currentAmmo"`
	LastShotUnixSec    int64             `json:"lastShotUnixSec"`
	IsCooling          bool              `json:"isCooling"`
	StringFatigue      float64           `json:"stringFatigue"`
	HistoryShots       []TimeSeriesPoint `json:"historyShots"`
}

type VirtualShootRequest struct {
	VariantCode string `json:"variantCode"`
	SessionID   string `json:"sessionId"`
	Mode        string `json:"mode"`
	BurstCount  int    `json:"burstCount"`
}

type VirtualShootResponse struct {
	SessionID string              `json:"sessionId"`
	ShotFired bool                `json:"shotFired"`
	Jammed    bool                `json:"jammed"`
	Recovered bool                `json:"recovered"`
	Message   string              `json:"message,omitempty"`
	NewState  *VirtualShootSession `json:"newState"`
}
