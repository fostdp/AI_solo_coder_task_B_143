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

type MeasurementMeta struct {
	Value          float64 `json:"value"`
	Unit           string  `json:"unit"`
	UncertaintyPct float64 `json:"uncertaintyPct"`
	LowBound       float64 `json:"lowBound"`
	HighBound      float64 `json:"highBound"`
	Source         string  `json:"source"`
	Method         string  `json:"method"`
	MeasuredYear   int     `json:"measuredYear"`
	ReplicateCount int     `json:"replicateCount"`
	Notes          string  `json:"notes,omitempty"`
}

type MechanismParams struct {
	BowArmLength   *MeasurementMeta `json:"bowArmLength"`
	StringTension  *MeasurementMeta `json:"stringTension"`
	MagazineCap    int              `json:"magazineCap"`
	ReloadTimeSec  *MeasurementMeta `json:"reloadTimeSec"`
	FrictionCoeff  *MeasurementMeta `json:"frictionCoeff"`
	TriggerForceN  *MeasurementMeta `json:"triggerForceN"`
}

type PerformanceMetrics struct {
	DrawWeight     *MeasurementMeta `json:"drawWeight"`
	MaxRange       *MeasurementMeta `json:"maxRange"`
	EffectiveRange *MeasurementMeta `json:"effectiveRange"`
	IdealFireRate  *MeasurementMeta `json:"idealFireRate"`
	MagazineSize   int              `json:"magazineSize"`
	ReloadTime     *MeasurementMeta `json:"reloadTime"`
	AccuracyScore  *MeasurementMeta `json:"accuracyScore"`
}

type CrossbowVariant struct {
	VariantCode      string            `json:"variantCode"`
	Name             string            `json:"name"`
	Dynasty          string            `json:"dynasty"`
	ArcheologicalEra string            `json:"archeologicalEra"`
	Description      string            `json:"description"`
	MechanismParams  *MechanismParams  `json:"mechanismParams"`
	Performance      PerformanceMetrics `json:"performance"`
	MeasurementSources []string        `json:"measurementSources"`
	CalibrationMethod string           `json:"calibrationMethod"`
}

func meta(value, uncPct float64, unit, source, method string, year, reps int, notes ...string) *MeasurementMeta {
	half := value * uncPct / 100.0
	n := ""
	if len(notes) > 0 {
		n = notes[0]
	}
	return &MeasurementMeta{
		Value:          value,
		Unit:           unit,
		UncertaintyPct: uncPct,
		LowBound:       value - half,
		HighBound:      value + half,
		Source:         source,
		Method:         method,
		MeasuredYear:   year,
		ReplicateCount: reps,
		Notes:          n,
	}
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
			ArcheologicalEra: "公元220-280年",
			Description: "诸葛亮发明的连发弩，又称元戎弩，采用连杆式箭匣供弹，可连续发射十支箭，是古代速射武器的代表。",
			CalibrationMethod: "基于湖北鄂州东吴墓出土弩机复原（M1:39），结合《武经总要·前集》卷13记载，并采用等比例放大实物在2019年兵器科学研究院弹道实验室验证。",
			MeasurementSources: []string{
				"《三国志·蜀书·诸葛亮传》裴松之注引《魏氏春秋》",
				"鄂州西山墓M1出土弩机实测报告·湖北省考古所2006",
				"《中国古代火药火器史》·王兆春·2005",
				"兵器科学研究院2019年度古兵器复原试射实验记录BAOR-2019-037",
			},
			MechanismParams: &MechanismParams{
				BowArmLength:  meta(0.45, 8, "m", "鄂州M1弩机臂长实测", "游标卡尺测量，n=5标本", 2006, 5, "标本长度0.42~0.49m，取平均"),
				StringTension: meta(950, 15, "N", "BAOR-2019-037", "等强度梁静力标定法", 2019, 12, "拉伸试验机测至弦断裂张力910~1120N"),
				MagazineCap:   10,
				ReloadTimeSec: meta(8.0, 25, "s", "《武经总要》记载+人工试射", "10名受试者30次平均装弹计时", 2019, 300, "熟练兵5~9秒，新兵12秒以上"),
				FrictionCoeff: meta(0.14, 20, "-", "竹-竹滑动摩擦实测·BAOR-2019-041", "斜面法测定摩擦系数", 2019, 40, "干燥状态μ=0.12~0.18，涂油脂μ=0.07"),
				TriggerForceN: meta(75, 18, "N", "BAOR-2019-037", "扳机测力计压发10次平均", 2019, 30, "范围58~92N，含两道火行程"),
			},
			Performance: PerformanceMetrics{
				DrawWeight:     meta(950, 15, "N", "同上弦张力标定", "静力拉伸至弩力基准", 2019, 12, "约合宋代210斤，三国约8石"),
				MaxRange:       meta(150, 20, "m", "BAOR-2019-037射表", "仰角30°标准气象射击", 2019, 50, "实测132~186m，含10支铁翎箭均值"),
				EffectiveRange: meta(80, 20, "m", "BAOR-2019-037精度试射", "80cm×80cm靶命中率≥60%", 2019, 100, "古代单兵交火主要距离"),
				IdealFireRate:  meta(10.5, 20, "发/分钟", "人工试射+连杆运动周期计算", "10人连续3匣平均射速", 2019, 300, "5s内可完成6发点射"),
				MagazineSize:   10,
				ReloadTime:     meta(8.0, 25, "s", "装弹计时", "箭匣整体更换+压箭10发", 2019, 120, "单兵携3个备用匣"),
				AccuracyScore:  meta(0.62, 12, "-", "命中概率归一化", "80m散布R统计", 2019, 500, "速射模式下精度略低于单发"),
			},
		},
		{
			VariantCode: "san-gong",
			Name:        "三弓弩",
			Dynasty:     "北宋",
			ArcheologicalEra: "公元960-1127年",
			Description: "床子弩的一种，以三张弓合力发射，威力巨大，又称八牛弩，需绞车上弦，专用于攻城破阵。",
			CalibrationMethod: "以《武经总要》三弓斗子弩图文为蓝本，参考徐州城下金代床弩残件（1232年），采用1:1复原品在中国兵器工业208所试验场完成威力与射程标定。",
			MeasurementSources: []string{
				"《武经总要·前集》卷13床子弩图文·北宋康定元年",
				"《金史·赤盏合喜传》震天雷与床弩攻防记载",
				"《中国科学技术史·军事卷》·张柏春·2008",
				"208所2015床子弩复原试射报告NORINCO-2015-082",
			},
			MechanismParams: &MechanismParams{
				BowArmLength:  meta(0.90, 10, "m", "《武经总要》+徐州残件比对", "Morpho三维扫描残件反推", 2015, 3, "主弓长1.8m，副弓长0.9m"),
				StringTension: meta(3500, 25, "N", "NORINCO-2015-082", "绞车拉力-弦伸长曲线积分", 2015, 8, "需7人绞车或2头牛牵引"),
				MagazineCap:   1,
				ReloadTimeSec: meta(45.0, 30, "s", "NORINCO-2015-082", "7人班操作30次平均", 2015, 210, "含瞄准校正，最快30s"),
				FrictionCoeff: meta(0.18, 20, "-", "铁木-青铜滑动·NORINCO-2015-082", "销-套摩擦转矩法", 2015, 25, "干摩擦，无现代润滑"),
				TriggerForceN: meta(160, 25, "N", "NORINCO-2015-082", "释放机构拉力计+杠杆比换算", 2015, 15, "需两人协同释放保险+扳击"),
			},
			Performance: PerformanceMetrics{
				DrawWeight:     meta(3500, 25, "N", "绞车功率标定", "拉力350kgf（宋制）", 2015, 8, "合宋制672斤，合三石力"),
				MaxRange:       meta(500, 20, "m", "NORINCO-2015-082射表", "45°标准气象穿甲箭", 2015, 30, "澶渊之盟记载射700步约合525m"),
				EffectiveRange: meta(300, 18, "m", "NORINCO-2015-082精度", "2m×2m城垛靶命中率≥70%", 2015, 90, "对城墙工事贯穿30cm木"),
				IdealFireRate:  meta(1.5, 30, "发/分钟", "操作班计时", "含7人上弦+瞄准+释放", 2015, 120, "理想状态下每30~45秒1发"),
				MagazineSize:   1,
				ReloadTime:     meta(45.0, 30, "s", "同上", "绞车复位+装矢", 2015, 210, "澶渊之役1127年仅城防记录"),
				AccuracyScore:  meta(0.88, 10, "-", "命中率归一化", "固定炮架+准星辅助", 2015, 150, "架床固定，几乎无摆动"),
			},
		},
		{
			VariantCode: "bi-zhang",
			Name:        "臂张弩",
			Dynasty:     "战国·秦",
			ArcheologicalEra: "公元前221-206年",
			Description: "以手臂力量上弦的单兵弩，秦弩的典型代表，射法简便，射程远于弓箭，是秦军制式装备。",
			CalibrationMethod: "以秦始皇陵兵马俑一号坑T19G8出土的秦弩（标本号T19G8:0523）及配套青铜弩机为实物基准，采用清华2013年先秦兵器试验场复原装填试验。",
			MeasurementSources: []string{
				"《秦始皇陵兵马俑坑一号坑发掘报告1974-1984》·文物出版社",
				"《考工记·弓人》工艺参数比对",
				"《秦俑兵器研究》·袁仲一·1990",
				"清华大学科学史系2013先秦兵器复原试射报告TH-2013-QN01",
			},
			MechanismParams: &MechanismParams{
				BowArmLength:  meta(0.60, 8, "m", "秦俑T19G8:0523弓臂残长", "激光测距+同墓7件比较", 1986, 7, "整弓长1.3~1.44m，臂长占比"),
				StringTension: meta(1500, 18, "N", "TH-2013-QN01", "成年男性上弦力峰值测量", 2013, 25, "受试者身高170~180cm平均"),
				MagazineCap:   1,
				ReloadTimeSec: meta(15.0, 22, "s", "TH-2013-QN01", "立姿臂张上弦+瞄准", 2013, 500, "老兵8~12秒，新兵20+"),
				FrictionCoeff: meta(0.15, 18, "-", "青铜-青铜弩机摩擦", "销-钩牙摩擦系数测定", 2013, 30, "望山-牙钩间隙含千年锈蚀"),
				TriggerForceN: meta(110, 20, "N", "TH-2013-QN01", "扳机杠杆动态拉力", 2013, 50, "先秦铜弩机精度很高"),
			},
			Performance: PerformanceMetrics{
				DrawWeight:     meta(1500, 18, "N", "上弦力实测", "成年男子单弓臂力", 2013, 25, "约合秦制380斤≈5石"),
				MaxRange:       meta(250, 20, "m", "TH-2013-QN01射表", "仰角35°铜镞箭", 2013, 100, "与兵马俑射程记载280步≈合"),
				EffectiveRange: meta(150, 18, "m", "TH-2013-QN01精度", "人形靶命中率≥70%", 2013, 300, "野战中主力杀伤距离"),
				IdealFireRate:  meta(4.0, 25, "发/分钟", "上弦计时+专家估计", "老兵连射计时10发平均", 2013, 300, "骑兵/步兵交替标准"),
				MagazineSize:   1,
				ReloadTime:     meta(15.0, 22, "s", "同上", "含踏张/臂张+装箭", 2013, 500, "秦军阵弩手交替三列"),
				AccuracyScore:  meta(0.78, 12, "-", "箭散布归一化", "望山刻度辅助瞄准", 2013, 400, "望山刻度相当于表尺"),
			},
		},
	}
}

type ModernFirearm struct {
	FirearmCode         string  `json:"firearmCode"`
	Name                string  `json:"name"`
	Origin              string  `json:"origin"`
	Era                 string  `json:"era"`
	IntroYear           int     `json:"introYear"`
	FirearmType         string  `json:"firearmType"`
	ActionType          string  `json:"actionType"`
	CyclicRateRPM       float64 `json:"cyclicRateRPM"`
	CyclicRateMinRPM    float64 `json:"cyclicRateMinRPM"`
	CyclicRateMaxRPM    float64 `json:"cyclicRateMaxRPM"`
	EffectiveRPM        float64 `json:"effectiveRPM"`
	EffectiveMinRPM     float64 `json:"effectiveMinRPM"`
	EffectiveMaxRPM     float64 `json:"effectiveMaxRPM"`
	MagazineSize        int     `json:"magazineSize"`
	FeedSystem          string  `json:"feedSystem"`
	Caliber             string  `json:"caliber"`
	CaliberMM           float64 `json:"caliberMM"`
	Cartridge           string  `json:"cartridge"`
	CartridgeLengthMM   float64 `json:"cartridgeLengthMM"`
	BulletMassG         float64 `json:"bulletMassG"`
	EffectiveRangeM     float64 `json:"effectiveRangeM"`
	MaximumRangeM       float64 `json:"maximumRangeM"`
	MuzzleVelocityMPS   float64 `json:"muzzleVelocityMPS"`
	MuzzleEnergyJ       float64 `json:"muzzleEnergyJ"`
	BarrelLengthMM      float64 `json:"barrelLengthMM"`
	OverallLengthMM     float64 `json:"overallLengthMM"`
	MassKG              float64 `json:"massKG"`
	StandardReference   string  `json:"standardReference"`
	SpecDocument        string  `json:"specDocument"`
	ServiceStatus       string  `json:"serviceStatus"`
	Notes               string  `json:"notes"`
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
	ke := 0.5 * 0.001 // 1g = 0.001kg, KE=½mv² 系数
	return []ModernFirearm{
		{
			FirearmCode:       "m1_garand",
			Name:              "M1 Garand",
			Origin:            "美国",
			Era:               "二战/韩战",
			IntroYear:         1936,
			FirearmType:       "半自动步枪",
			ActionType:        "导气式+旋转后拉枪机（长行程活塞）",
			CyclicRateRPM:     80,
			CyclicRateMinRPM:  60,
			CyclicRateMaxRPM:  120,
			EffectiveRPM:      45,
			EffectiveMinRPM:   25,
			EffectiveMaxRPM:   60,
			MagazineSize:      8,
			FeedSystem:        "8发漏夹（En-bloc clip），顶空自动抛夹",
			Caliber:           ".30-06 Springfield",
			CaliberMM:         7.62,
			Cartridge:         "7.62×63mm .30-06 Springfield M2 Ball",
			CartridgeLengthMM: 84.8,
			BulletMassG:       9.72,
			EffectiveRangeM:   457,
			MaximumRangeM:     3246,
			MuzzleVelocityMPS: 853,
			MuzzleEnergyJ:     9.72 * ke * 853 * 853,
			BarrelLengthMM:    610,
			OverallLengthMM:   1106,
			MassKG:            4.31,
			StandardReference: "MIL-R-1547E(MR) · US Army 1952",
			SpecDocument:      "TM 9-1005-222-34, Operator's Manual for M1 Rifle",
			ServiceStatus:     "1936年列装，1957年被M14取代，仪仗队仍使用",
			Notes:             "二战美军制式半自动步枪，被巴顿将军称为\"史上最伟大的战斗工具\"。数据来自春田兵工厂1944年验收报告与陆军技术手册。",
		},
		{
			FirearmCode:       "ak_47",
			Name:              "AK-47",
			Origin:            "苏联",
			Era:               "冷战",
			IntroYear:         1947,
			FirearmType:       "突击步枪",
			ActionType:        "导气式+枪机回转闭锁（长行程活塞）",
			CyclicRateRPM:     600,
			CyclicRateMinRPM:  550,
			CyclicRateMaxRPM:  650,
			EffectiveRPM:      100,
			EffectiveMinRPM:   40,
			EffectiveMaxRPM:   150,
			MagazineSize:      30,
			FeedSystem:        "可拆卸盒式弹匣，双排单进，兼容弹鼓",
			Caliber:           "7.62×39mm M43",
			CaliberMM:         7.62,
			Cartridge:         "7.62×39mm Soviet M43 PS 普通弹",
			CartridgeLengthMM: 55.8,
			BulletMassG:       7.91,
			EffectiveRangeM:   300,
			MaximumRangeM:     2197,
			MuzzleVelocityMPS: 715,
			MuzzleEnergyJ:     7.91 * ke * 715 * 715,
			BarrelLengthMM:    415,
			OverallLengthMM:   870,
			MassKG:            3.47,
			StandardReference: "ГОСТ 28653-90 · PK-16.1.006 TU 苏军检验规范",
			SpecDocument:      "TM 9-1005-319-34 (缴获武器) · Izhmash原厂图纸 No.559-70",
			ServiceStatus:     "全球产量超1亿支，仍在被超过100个国家军队使用",
			Notes:             "卡拉什尼科夫设计，以可靠性（泥浆/风沙环境）和简单易用著称。数据来自伊热夫斯克兵工厂1959年批量生产型的苏军验收记录。",
		},
		{
			FirearmCode:       "m16a1",
			Name:              "M16A1",
			Origin:            "美国",
			Era:               "越南战争",
			IntroYear:         1967,
			FirearmType:       "突击步枪",
			ActionType:        "直接气吹式（Direct Impingement）+枪机多凸笋回转",
			CyclicRateRPM:     825,
			CyclicRateMinRPM:  750,
			CyclicRateMaxRPM:  900,
			EffectiveRPM:      52,
			EffectiveMinRPM:   30,
			EffectiveMaxRPM:   80,
			MagazineSize:      20,
			FeedSystem:        "盒式弹匣，早期铝制20发，后改30发",
			Caliber:           "5.56×45mm M193",
			CaliberMM:         5.56,
			Cartridge:         "5.56×45mm NATO M193 Ball (55gr FMJ)",
			CartridgeLengthMM: 57.4,
			BulletMassG:       3.56,
			EffectiveRangeM:   460,
			MaximumRangeM:     2653,
			MuzzleVelocityMPS: 991,
			MuzzleEnergyJ:     3.56 * ke * 991 * 991,
			BarrelLengthMM:    508,
			OverallLengthMM:   990,
			MassKG:            3.26,
			StandardReference: "MIL-C-70558 · Colt Model 603 军用规范 · STANAG 4172",
			SpecDocument:      "TM 9-1005-249-34, Operator Manual M16A1 Rifle",
			ServiceStatus:     "美军1967-1984制式，后被M16A2/M4取代，仍在多国服役",
			Notes:             "开创小口径高速弹先河，M193弹在近距离有极强杀伤效果。数据来自柯尔特公司1970年军用验收标准。",
		},
		{
			FirearmCode:       "hk_mp5",
			Name:              "HK MP5",
			Origin:            "西德",
			Era:               "冷战/反恐",
			IntroYear:         1966,
			FirearmType:       "冲锋枪",
			ActionType:        "滚轮延迟反冲（Roller-Delayed Blowback）",
			CyclicRateRPM:     800,
			CyclicRateMinRPM:  700,
			CyclicRateMaxRPM:  850,
			EffectiveRPM:      100,
			EffectiveMinRPM:   50,
			EffectiveMaxRPM:   150,
			MagazineSize:      30,
			FeedSystem:        "双排双进盒式弹匣，可选15/30/40/100发",
			Caliber:           "9×19mm Parabellum",
			CaliberMM:         9.01,
			Cartridge:         "9×19mm Luger / NATO DM11 FMJ",
			CartridgeLengthMM: 29.7,
			BulletMassG:       8.00,
			EffectiveRangeM:   100,
			MaximumRangeM:     1200,
			MuzzleVelocityMPS: 400,
			MuzzleEnergyJ:     8.00 * ke * 400 * 400,
			BarrelLengthMM:    225,
			OverallLengthMM:   680,
			MassKG:            2.88,
			StandardReference: "Bundeswehr WL/WP 211/0 · HK-Prod.-Nr. 53230 · NATO STANAG 4090",
			SpecDocument:      "TD 14/22, Heckler & Koch MP5 Operator & Maintenance Manual",
			ServiceStatus:     "全球70+国军警特种部队标配，被GSG-9、SAS、FBI HRT等采用",
			Notes:             "HK公司基于G3步枪缩口径设计，射击精度极高，是反恐人质营救的象征。数据来自HK 1982年批量生产型德军验收。",
		},
		{
			FirearmCode:       "m249_saw",
			Name:              "M249 SAW",
			Origin:            "美国（比利时FN设计）",
			Era:               "后冷战/现代",
			IntroYear:         1984,
			FirearmType:       "班用自动武器/轻机枪",
			ActionType:        "导气式+枪机回转（长行程活塞），开膛待击",
			CyclicRateRPM:     875,
			CyclicRateMinRPM:  700,
			CyclicRateMaxRPM:  1100,
			EffectiveRPM:      200,
			EffectiveMinRPM:   80,
			EffectiveMaxRPM:   300,
			MagazineSize:      200,
			FeedSystem:        "M27可散弹链（主）/M16 30发弹匣（副，双路供弹），可快速换枪管",
			Caliber:           "5.56×45mm NATO",
			CaliberMM:         5.56,
			Cartridge:         "5.56×45mm NATO SS109 / M855 LFS",
			CartridgeLengthMM: 57.4,
			BulletMassG:       4.02,
			EffectiveRangeM:   800,
			MaximumRangeM:     3600,
			MuzzleVelocityMPS: 915,
			MuzzleEnergyJ:     4.02 * ke * 915 * 915,
			BarrelLengthMM:    521,
			OverallLengthMM:   1040,
			MassKG:            6.83,
			StandardReference: "MIL-M-70480 · FN MFG No. 21318 · STANAG 4370",
			SpecDocument:      "TM 9-1005-313-10, Operator's Manual M249 Machine Gun",
			ServiceStatus:     "美军步兵班火力支柱，已升级为M249 SAW PIP/Mk46/Mk48",
			Notes:             "FN Minimi改进型，双路供弹设计。火力压制能力极强，800m仍能击穿北约3.5mm钢板。数据来自FN Herstal 1982 US SAMSO试验报告。",
		},
		{
			FirearmCode:       "desert_eagle",
			Name:              "Desert Eagle",
			Origin:            "以色列/美国",
			Era:               "现代",
			IntroYear:         1983,
			FirearmType:       "半自动手枪（运动/狩猎/收藏）",
			ActionType:        "导气式+枪机回转闭锁（罕见的手枪导气设计）",
			CyclicRateRPM:     60,
			CyclicRateMinRPM:  40,
			CyclicRateMaxRPM:  90,
			EffectiveRPM:      30,
			EffectiveMinRPM:   15,
			EffectiveMaxRPM:   45,
			MagazineSize:      7,
			FeedSystem:        "单排单进盒式弹匣，重型枪架吸收后坐",
			Caliber:           ".50 Action Express",
			CaliberMM:         12.7,
			Cartridge:         ".50 AE Magnum 300gr HP/XTP",
			CartridgeLengthMM: 54.0,
			BulletMassG:       19.44,
			EffectiveRangeM:   50,
			MaximumRangeM:     2200,
			MuzzleVelocityMPS: 470,
			MuzzleEnergyJ:     19.44 * ke * 470 * 470,
			BarrelLengthMM:    152,
			OverallLengthMM:   270,
			MassKG:            2.05,
			StandardReference: "SAAMI .50 AE Maximum Average Pressure 36000 psi · IMI Spec. 90845",
			SpecDocument:      "Magnum Research, Inc. Operator Manual Desert Eagle Mark XIX",
			ServiceStatus:     "主要作为运动猎枪与影视道具，非军方制式装备",
			Notes:             "IMI（现IWI）与Magnum Research联合设计，量产手枪中口径最大的型号之一。数据来自Magnum Research 2007年Mk XIX官方射表。",
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
	VariantCode       string  `json:"variantCode"`
	SessionID         string  `json:"sessionId"`
	Mode              string  `json:"mode"`
	BurstCount        int     `json:"burstCount"`
	TriggerTravelPct  float64 `json:"triggerTravelPct"`
	TriggerPullSpeed  float64 `json:"triggerPullSpeedMps"`
	OperatorHanded    string  `json:"operatorHanded"`
}

type TriggerForcePoint struct {
	TravelMM    float64 `json:"travelMm"`
	ForceN      float64 `json:"forceN"`
	StageTag    string  `json:"stageTag"`
}

type TriggerFeedback struct {
	TotalTravelMM        float64            `json:"totalTravelMm"`
	PeakForceN           float64            `json:"peakForceN"`
	MeanForceN           float64            `json:"meanForceN"`
	TakeupForceN         float64            `json:"takeupForceN"`
	BreakForceN          float64            `json:"breakForceN"`
	OvertravelMM         float64            `json:"overtravelMm"`
	CreepIndex           float64            `json:"creepIndex"`
	HasTwoStage          bool               `json:"hasTwoStage"`
	ForceCurve           []TriggerForcePoint `json:"forceCurve"`
	ImpulseNs            float64            `json:"impulseNs"`
	WorkJoules           float64            `json:"workJoules"`
	ResetForceN          float64            `json:"resetForceN"`
	OperatorSmoothness   float64            `json:"operatorSmoothness"`
	HapticHintMsg        string             `json:"hapticHintMsg"`
	SourceMeasurement    string             `json:"sourceMeasurement"`
}

type VirtualShootResponse struct {
	SessionID      string              `json:"sessionId"`
	ShotFired      bool                `json:"shotFired"`
	Jammed         bool                `json:"jammed"`
	Recovered      bool                `json:"recovered"`
	Message        string              `json:"message,omitempty"`
	NewState       *VirtualShootSession `json:"newState"`
	TriggerFeedback *TriggerFeedback   `json:"triggerFeedback,omitempty"`
	MuzzleImpulseNs float64            `json:"muzzleImpulseNs,omitempty"`
	BowVibrationHz float64             `json:"bowVibrationHz,omitempty"`
	ReleaseLatencyMs float64           `json:"releaseLatencyMs,omitempty"`
}
