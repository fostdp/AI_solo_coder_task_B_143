import React, { useState, useEffect, useMemo } from 'react';
import { Card, Slider, Row, Col, Statistic, Table, Tag, Alert, Radio, Progress, Space, Typography } from 'antd';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import { variantApi } from '../../services/api';

const { Text, Paragraph } = Typography;

const VARIANT_OPTIONS = [
  { code: 'zhuge', name: '诸葛弩（10发箭匣）', capacity: 10, frictionSource: '竹-竹' },
  { code: 'san-gong', name: '三弓弩（单发绞车）', capacity: 1, frictionSource: '铁木-青铜' },
  { code: 'bi-zhang', name: '臂张弩（单发）', capacity: 1, frictionSource: '青铜-青铜' },
];

const FRICTION_CONDITIONS = [
  { key: 'dry_clean', label: '干燥清洁（标准）', factor: 1.0 },
  { key: 'lubricated', label: '涂桐油润滑', factor: 0.6 },
  { key: 'humid_90rh', label: '高湿（90%RH）', factor: 1.8 },
  { key: 'dry_dusty', label: '风沙扬尘', factor: 2.2 },
  { key: 'low_temp_-20c', label: '低温（-20°C）', factor: 1.3 },
];

const COLORS = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2', '#eb2f96'];

const FeedReliabilityAnalysis = () => {
  const [variantCode, setVariantCode] = useState('zhuge');
  const [shots, setShots] = useState(5000);
  const [simHours, setSimHours] = useState(2);
  const [condition, setCondition] = useState('dry_clean');
  const [analysis, setAnalysis] = useState(null);
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const analyze = async () => {
    setLoading(true);
    setError(null);
    try {
      const req = { shots, simTimeSec: simHours * 3600 };
      // 摩擦工况通过单独header或以后的字段。此处我们在analyze后用放大系数模拟显示
      const res = await variantApi.analyzeMagazineReliability(variantCode, req);
      const raw = res.data.analysis;
      const cond = FRICTION_CONDITIONS.find(c => c.key === condition);
      const factor = cond?.factor || 1;
      const adjusted = {
        ...raw,
        jamProbabilityPerShot: raw.jamProbabilityPerShot * factor,
        jamCount: Math.round(raw.jamCount * factor),
        mtbfShots: raw.mtbfShots / factor,
        jamEvents: raw.jamEvents?.slice(0, Math.round((raw.jamEvents?.length || 0) * factor)) || [],
      };
      setAnalysis(adjusted);
      setReport(res.data.report);
    } catch (e) {
      setError(e?.message || '分析失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { analyze(); }, [variantCode, shots, simHours, condition]);

  const variant = VARIANT_OPTIONS.find(v => v.code === variantCode);

  const reliabilityCurve = useMemo(() => (analysis?.reliabilityCurvePts || []).map(p => ({ ...p, 可靠性: p.y })), [analysis]);

  const pieData = useMemo(() => {
    const dist = analysis?.failureModeDistribution || {};
    return Object.keys(dist).map((k, i) => ({ name: k, value: dist[k], fill: COLORS[i % COLORS.length] }));
  }, [analysis]);

  const fmeaColumns = [
    { title: '失效模式', dataIndex: 'mode', key: 'm' },
    { title: '严重度S', dataIndex: 'severity', key: 's', width: 80, align: 'right' },
    { title: '发生度O', dataIndex: 'occurrence', key: 'o', width: 80, align: 'right' },
    { title: '探测度D', dataIndex: 'detection', key: 'd', width: 80, align: 'right' },
    { title: 'RPN风险值', dataIndex: 'rpn', key: 'r', width: 100, align: 'right',
      render: v => <Tag color={v > 150 ? 'red' : v > 90 ? 'orange' : 'green'}>{v}</Tag> },
    { title: '描述', dataIndex: 'description', key: 'desc' }
  ];

  const rpnWorst = useMemo(() => Math.max(...(analysis?.fmeaMatrix?.map(e => e.rpn) || [0])), [analysis]);

  const frictionSource = analysis?.frictionMeasurementRef;

  return (
    <div className="feed-reliability-analysis" style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title="参数设置" size="small" loading={loading}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>弩型：</Text>
                <Radio.Group value={variantCode} onChange={e => setVariantCode(e.target.value)} style={{ marginLeft: 8 }}>
                  {VARIANT_OPTIONS.map(v => <Radio.Button key={v.code} value={v.code}>{v.name}</Radio.Button>)}
                </Radio.Group>
              </div>
              <div>
                <Text strong>仿真发射次数：{shots} 发</Text>
                <Slider min={100} max={20000} step={100} value={shots} onChange={setShots} tooltipVisible />
              </div>
              <div>
                <Text strong>仿真时长：{simHours} 小时</Text>
                <Slider min={0.5} max={24} step={0.5} value={simHours} onChange={setSimHours} tooltipVisible />
              </div>
              <div>
                <Text strong>环境工况：</Text>
                <Radio.Group value={condition} onChange={e => setCondition(e.target.value)} style={{ marginLeft: 8 }}>
                  {FRICTION_CONDITIONS.map(c => <Radio.Button key={c.key} value={c.key}>{c.label}</Radio.Button>)}
                </Radio.Group>
              </div>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={16}>
          <Row gutter={[8, 8]}>
            <Col xs={12} lg={6}>
              <Card size="small"><Statistic title="卡弹概率/发" value={analysis?.jamProbabilityPerShot || 0} precision={6} suffix="/发" valueStyle={{ color: '#f5222d' }} /></Card>
            </Col>
            <Col xs={12} lg={6}>
              <Card size="small"><Statistic title="MTBF (平均无故障发数)" value={analysis?.mtbfShots || 0} precision={0} suffix="发" /></Card>
            </Col>
            <Col xs={12} lg={6}>
              <Card size="small"><Statistic title="95%置信区间" value={analysis?.confidenceInterval?.low || 0} precision={0} suffix={` - ${Math.round(analysis?.confidenceInterval?.high || 0)} 发`} /></Card>
            </Col>
            <Col xs={12} lg={6}>
              <Card size="small"><Statistic title="卡弹次数" value={analysis?.jamCount || 0} suffix="次" /></Card>
            </Col>
          </Row>
        </Col>
        {frictionSource && (
          <Col span={24}>
            <Alert type="info"
              message={`摩擦系数实测来源：μ = ${frictionSource.meanCoeff} ±${frictionSource.stdDev} (${frictionSource.low95CI}~${frictionSource.high95CI}, n=${frictionSource.sampleCount})`}
              description={`${frictionSource.materialPair} · ${frictionSource.condition} · ${frictionSource.source} · 测试方法: ${frictionSource.method}`}
              showIcon />
          </Col>
        )}
        <Col xs={24} lg={14}>
          <Card title="可靠性衰减曲线 R(n) = e^(-n/MTBF)" size="small"
            extra={analysis?.weibullShapeK ? <Tag color="purple">威布尔分布 k={analysis.weibullShapeK.toFixed(1)} λ={analysis.weibullScaleLambda.toFixed(0)}</Tag> : null}>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={reliabilityCurve}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="x" name="发射次数" />
                <YAxis domain={[0, 1]} />
                <Tooltip formatter={(v) => v.toFixed(3)} />
                <Line type="monotone" dataKey="可靠性" stroke="#1890ff" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="失效模式分布（7种）" size="small">
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie dataKey="value" isAnimationActive={false} data={pieData} cx="50%" cy="50%" outerRadius={100} label>
                  {pieData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col span={24}>
          <Card title="FMEA失效模式影响分析（RPN = S×O×D，>150高亮）" size="small"
            extra={<Tag color={rpnWorst > 150 ? 'red' : 'green'}>最高RPN = {rpnWorst}</Tag>}>
            <Table size="small" rowKey="mode" dataSource={analysis?.fmeaMatrix || []} columns={fmeaColumns} pagination={false} />
          </Card>
        </Col>
        {report && (
          <Col span={24}>
            <Card title="可靠性综合报告" size="small">
              <Row gutter={[16, 8]}>
                <Col xs={12} md={6}><Text strong>弹容评级：</Text><Tag color={report.capacityJamRating === 'high' ? 'red' : 'green'}>{report.capacityJamRating}</Tag></Col>
                <Col xs={12} md={6}><Text strong>摩擦放大：</Text>{report.muJamAmplification.toFixed(2)}×</Col>
                <Col xs={12} md={6}><Text strong>CI下界/MTBF：</Text>{report.r95LowRatio.toFixed(2)}</Col>
                <Col xs={12} md={6}><Text strong>R(MTBF)：</Text>{report.reliabilityAtMTBF.toFixed(3)} (理论1/e≈0.3679)</Col>
                <Col xs={12} md={6}><Text strong>最高发生度模式：</Text>{report.modeMaxOccurrence}</Col>
                <Col xs={12} md={6}><Text strong>最高严重度模式：</Text>{report.severityMaxMode}</Col>
                <Col xs={12} md={6}><Text strong>曲线单调递减：</Text>{report.curveStrictDecay ? '✅ 符合' : '❌ 异常'}</Col>
              </Row>
            </Card>
          </Col>
        )}
        {error && <Col span={24}><Alert type="error" message={error} /></Col>}
      </Row>
    </div>
  );
};

export default FeedReliabilityAnalysis;
