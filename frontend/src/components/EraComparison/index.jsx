import React, { useState, useEffect, useMemo } from 'react';
import { Table, Card, Select, Statistic, Row, Col, Tag, Alert, Typography } from 'antd';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, ScatterChart, Scatter, ZAxis } from 'recharts';
import { firearmApi } from '../../services/api';

const { Option } = Select;
const { Text } = Typography;

const ANCIENT_OPTIONS = [
  { code: 'zhuge', name: '诸葛弩', era: '三国' },
  { code: 'san-gong', name: '三弓弩', era: '北宋' },
  { code: 'bi-zhang', name: '臂张弩', era: '战国秦' },
];

const MODERN_OPTIONS = [
  { code: 'm1_garand', name: 'M1 Garand', year: 1936, type: '半自动' },
  { code: 'ak_47', name: 'AK-47', year: 1947, type: '突击步枪' },
  { code: 'm16a1', name: 'M16A1', year: 1967, type: '突击步枪' },
  { code: 'hk_mp5', name: 'HK MP5', year: 1966, type: '冲锋枪' },
  { code: 'm249_saw', name: 'M249 SAW', year: 1984, type: '轻机枪' },
  { code: 'desert_eagle', name: 'Desert Eagle', year: 1983, type: '手枪' },
];

const GAP_KPIS = [
  { key: 'fireRate', label: '射速差距', ancientUnit: '发/分', modernUnit: '发/分', color: '#f5222d' },
  { key: 'effectiveRange', label: '有效射程差距', ancientUnit: 'm', modernUnit: 'm', color: '#fa8c16' },
  { key: 'magazineSize', label: '弹容差距', ancientUnit: '发', modernUnit: '发', color: '#faad14' },
  { key: 'muzzleVelocity', label: '初速差距', ancientUnit: 'm/s', modernUnit: 'm/s', color: '#389e0d' },
];

const EraComparison = ({ ancientVariants = null, modernFirearms = null }) => {
  const [ancient, setAncient] = useState(ancientVariants || []);
  const [modern, setModern] = useState(modernFirearms || []);
  const [selectedAncient, setSelectedAncient] = useState(['zhuge', 'san-gong', 'bi-zhang']);
  const [selectedModern, setSelectedModern] = useState(['ak_47', 'm249_saw', 'hk_mp5']);
  const [comparison, setComparison] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!modernFirearms) {
      firearmApi.listModernFirearms().then(r => setModern(r.data.items)).catch(e => setError(e.message));
    }
  }, [modernFirearms]);

  const handleCompare = async () => {
    setLoading(true);
    setError(null);
    try {
      const req = {
        ancientVariants: selectedAncient,
        modernFirearms: selectedModern,
        compareMetrics: GAP_KPIS.map(k => k.key)
      };
      // 支持双key（name/code）查询
      const reqModern = selectedModern.map(c => {
        const found = MODERN_OPTIONS.find(o => o.code === c);
        return found ? found.name : c;
      });
      req.modernFirearms = reqModern;
      const res = await firearmApi.compareEraFirearms(req);
      setComparison(res.data);
    } catch (e) {
      setError(e?.message || '对比失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { if (selectedAncient.length && selectedModern.length) handleCompare(); }, [selectedAncient, selectedModern]);

  const bestAncient = useMemo(() => comparison?.bestAncient || { fireRate: 0, effectiveRange: 0, magazineSize: 0 }, [comparison]);
  const bestModern = useMemo(() => comparison?.bestModern || { fireRate: 0, effectiveRange: 0, magazineSize: 0, muzzleVelocity: 0 }, [comparison]);

  const gapTable = useMemo(() => comparison?.eraGapTable || [], [comparison]);

  const barData = useMemo(() => {
    const g = gapTable;
    return GAP_KPIS.map((k, i) => g[i] ? {
      metric: k.label,
      古代: g[i].ancientValue,
      现代: g[i].modernValue,
      倍数: g[i].gapRatio.toFixed(1)
    } : null).filter(Boolean);
  }, [gapTable]);

  const timelineData = useMemo(() => [
    ...ANCIENT_OPTIONS.map(a => ({ name: a.name, year: a.era === '三国' ? 230 : a.era === '北宋' ? 1000 : -220, type: '古代', rpm: a.code === 'zhuge' ? 10.5 : a.code === 'bi-zhang' ? 4 : 1.5 })),
    ...MODERN_OPTIONS.map(m => ({ name: m.name, year: m.year, type: '现代', rpm: m.code === 'm249_saw' ? 200 : m.code === 'ak_47' ? 100 : m.code === 'm16a1' ? 52 : m.code === 'hk_mp5' ? 100 : m.code === 'm1_garand' ? 45 : 30 }))
  ], []);

  const modernAncientRatio = useMemo(() => {
    const avgModern = (bestModern.fireRate + bestModern.effectiveRange + bestModern.magazineSize) / 3;
    const avgAncient = (bestAncient.fireRate + bestAncient.effectiveRange + bestAncient.magazineSize) / 3;
    return avgModern / Math.max(1e-6, avgAncient);
  }, [bestAncient, bestModern]);

  const columns = [
    { title: '对比指标', dataIndex: 'metric', key: 'm' },
    { title: '古代最优', dataIndex: 'ancientValue', key: 'a', render: (v, r) => `${v} ${r.ancientUnit}`, align: 'right' },
    { title: '现代最优', dataIndex: 'modernValue', key: 'm2', render: (v, r) => `${v} ${r.modernUnit}`, align: 'right' },
    { title: '技术差距倍数', dataIndex: 'gapRatio', key: 'r', render: v => <Tag color={v > 10 ? 'red' : v > 5 ? 'orange' : 'blue'}>{v.toFixed(2)}×</Tag>, align: 'right' },
    { title: '说明', dataIndex: 'remark', key: 'rm', ellipsis: true }
  ];

  return (
    <div className="era-comparison" style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="古代武器选择" size="small" extra={<Text type="secondary">多选对比</Text>}>
            <Select mode="multiple" style={{ width: '100%' }} value={selectedAncient} onChange={setSelectedAncient} placeholder="选择古代弩型">
              {ANCIENT_OPTIONS.map(a => <Option key={a.code} value={a.code}>{a.name} ({a.era})</Option>)}
            </Select>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="现代武器选择" size="small" extra={<Text type="secondary">多选对比</Text>}>
            <Select mode="multiple" style={{ width: '100%' }} value={selectedModern} onChange={setSelectedModern} placeholder="选择现代枪械">
              {MODERN_OPTIONS.map(m => <Option key={m.code} value={m.code}>{m.name} ({m.year}, {m.type})</Option>)}
            </Select>
          </Card>
        </Col>
        {error && <Col span={24}><Alert type="error" message={error} /></Col>}
        {GAP_KPIS.map((k, i) => {
          const entry = gapTable[i];
          return (
            <Col xs={12} lg={6} key={k.key}>
              <Card size="small" style={{borderTop: `3px solid ${k.color}`}}>
                <Statistic title={k.label} value={entry?.gapRatio || 0} precision={1} suffix="×" valueStyle={{ color: k.color }} />
                <div style={{ fontSize: 11, color: '#666', marginTop: 4 }}>
                  {entry?.ancientValue}{k.ancientUnit} → {entry?.modernValue}{k.modernUnit}
                </div>
              </Card>
            </Col>
          );
        })}
        <Col span={24}>
          <Card title="技术进步时间线（红点为古代，蓝点为现代）" size="small" extra={<Tag color="magenta">综合差距≈{modernAncientRatio.toFixed(0)}倍</Tag>}>
            <ResponsiveContainer width="100%" height={280}>
              <ScatterChart margin={{ top: 20, right: 20, bottom: 20, left: 20 }}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="year" name="年份" tick={{fontSize:10}} />
                <YAxis dataKey="rpm" name="射速RPM" />
                <ZAxis dataKey="rpm" range={[50, 500]} />
                <Tooltip cursor={{ strokeDasharray: '3 3' }} />
                <Legend />
                <Scatter name="古代" data={timelineData.filter(d => d.type === '古代')} fill="#f5222d" />
                <Scatter name="现代" data={timelineData.filter(d => d.type === '现代')} fill="#1890ff" />
              </ScatterChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card title="古代-现代双柱对比" size="small">
            <ResponsiveContainer width="100%" height={340}>
              <BarChart data={barData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="metric" />
                <YAxis yAxisId="left" />
                <YAxis yAxisId="right" orientation="right" />
                <Tooltip />
                <Legend />
                <Bar yAxisId="left" dataKey="古代" fill="#faad14" />
                <Bar yAxisId="left" dataKey="现代" fill="#1890ff" />
                <Bar yAxisId="right" dataKey="倍数" fill="#52c41a" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="技术差距表" size="small" loading={loading}>
            <Table size="small" rowKey="metric" dataSource={gapTable} columns={columns} pagination={false} />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default EraComparison;
