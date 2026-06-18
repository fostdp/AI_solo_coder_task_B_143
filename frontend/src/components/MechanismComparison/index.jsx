import React, { useState, useEffect, useMemo } from 'react';
import { Table, Card, Checkbox, Space, Button, Tag, Statistic, Row, Col, Alert, Tooltip } from 'antd';
import { RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar, Legend, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import { variantApi } from '../../services/api';

const { Meta } = Card;

const ALL_METRICS = [
  { key: 'idealFireRate', label: '射速(发/分)', higherIsBetter: true },
  { key: 'maxRange', label: '最大射程(m)', higherIsBetter: true },
  { key: 'effectiveRange', label: '有效射程(m)', higherIsBetter: true },
  { key: 'drawWeight', label: '张力(N)', higherIsBetter: false },
  { key: 'reloadTime', label: '装填时间(s)', higherIsBetter: false },
  { key: 'magazineSize', label: '弹容(发)', higherIsBetter: true },
  { key: 'accuracyScore', label: '精度评分(0-1)', higherIsBetter: true },
];

const MEASUREMENT_SOURCES = {
  zhuge: ['《三国志》裴松之注', '鄂州M1出土弩机', 'BAOR-2019-037复原试射'],
  'san-gong': ['《武经总要》床子弩图文', '徐州城下残件', 'NORINCO-2015-082'],
  'bi-zhang': ['秦俑一号坑弩机T19G8:0523', '《考工记·弓人》', 'TH-2013-QN01'],
};

const MechanismComparison = ({ variants = null, compareMetrics = null, showSources = true, onCompare = null }) => {
  const [variantList, setVariantList] = useState([]);
  const [selectedCodes, setSelectedCodes] = useState(['zhuge', 'san-gong', 'bi-zhang']);
  const [metrics, setMetrics] = useState(compareMetrics || ALL_METRICS.map(m => m.key));
  const [comparison, setComparison] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (variants) {
      setVariantList(variants);
    } else {
      variantApi.listVariants().then(r => setVariantList(r.data.items)).catch(e => setError(e.message));
    }
  }, [variants]);

  const selectedVariants = useMemo(() =>
    variantList.filter(v => selectedCodes.includes(v.variantCode)), [variantList, selectedCodes]);

  const getMetaVal = (m, metric) => {
    const p = m?.performance;
    if (!p) return 0;
    const map = { idealFireRate: p.idealFireRate, maxRange: p.maxRange, effectiveRange: p.effectiveRange, drawWeight: p.drawWeight, reloadTime: p.reloadTime, magazineSize: { value: p.magazineSize }, accuracyScore: p.accuracyScore };
    const node = map[metric];
    return typeof node === 'object' && node ? node.value : (node ?? 0);
  };

  const getMetaUnc = (m, metric) => {
    const p = m?.performance;
    if (!p) return null;
    const map = { idealFireRate: p.idealFireRate, maxRange: p.maxRange, effectiveRange: p.effectiveRange, drawWeight: p.drawWeight, reloadTime: p.reloadTime, accuracyScore: p.accuracyScore };
    return map[metric] || null;
  };

  const handleCompare = async () => {
    setLoading(true);
    setError(null);
    try {
      let res;
      if (onCompare) {
        res = await onCompare(selectedCodes, metrics);
      } else {
        res = await variantApi.compareVariants({ variantCodes: selectedCodes, compareMetrics: metrics });
      }
      setComparison(res.data);
    } catch (e) {
      setError(e?.message || '对比失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { if (selectedVariants.length >= 2) handleCompare(); }, [selectedCodes.length]);

  const radarData = useMemo(() => {
    if (!comparison?.performanceRadar) return [];
    return comparison.performanceRadar.map(r => {
      const obj = { metric: r.metric };
      Object.keys(r.values || {}).forEach(k => obj[k] = r.values[k]);
      return obj;
    });
  }, [comparison]);

  const computeAdvantage = () => {
    if (!comparison?.advantageMap) return [];
    return comparison.advantageMap.map(a => ({
      ...a,
      isGood: a.advantageRatio >= 1.5
    }));
  };

  const columns = [
    { title: '弩型', dataIndex: 'name', key: 'name', width: 120, render: (t, r) => <strong>{t}<div style={{fontSize:11,color:'#888'}}>{r.dynasty}</div></strong> },
    ...ALL_METRICS.filter(m => metrics.includes(m.key)).map(m => ({
      title: m.label, dataIndex: m.key, key: m.key, width: 110, align: 'right',
      render: (_, rec) => {
        const val = getMetaVal(rec, m.key);
        const unc = getMetaUnc(rec, m.key);
        const best = m.higherIsBetter
          ? Math.max(...selectedVariants.map(v => getMetaVal(v, m.key)))
          : Math.min(...selectedVariants.map(v => getMetaVal(v, m.key)));
        const isBest = Math.abs(val - best) < 1e-6;
        return (
          <Tooltip title={unc?.source ? `来源: ${unc.source}\n不确定度: ±${unc.uncertaintyPct}%` : '无实验数据'}>
            <span style={{ color: isBest ? '#1890ff' : '#333', fontWeight: isBest ? 600 : 400 }}>
              {typeof val === 'number' ? val.toFixed(2) : val}
              {unc?.uncertaintyPct ? <span style={{fontSize:10,color:'#999',marginLeft:2}}>±{unc.uncertaintyPct}%</span> : null}
            </span>
          </Tooltip>
        );
      }
    })),
    { title: '实验来源', key: 'src', width: 200, render: (_, rec) => showSources && MEASUREMENT_SOURCES[rec.variantCode]?.map((s, i) => <Tag key={i} color="geekblue" style={{marginBottom:4}}>{s}</Tag>) }
  ];

  return (
    <div className="mechanism-comparison" style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="选择弩型与指标" size="small" extra={<Button type="primary" loading={loading} onClick={handleCompare}>对比分析</Button>}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <div style={{ marginBottom: 6, fontWeight: 600 }}>弩型选择：</div>
                <Checkbox.Group value={selectedCodes} onChange={setSelectedCodes}>
                  {variantList.map(v => (
                    <Checkbox key={v.variantCode} value={v.variantCode}>{v.name}（{v.dynasty}）</Checkbox>
                  ))}
                </Checkbox.Group>
              </div>
              <div>
                <div style={{ marginBottom: 6, fontWeight: 600 }}>性能指标：</div>
                <Checkbox.Group value={metrics} onChange={setMetrics}>
                  {ALL_METRICS.map(m => (
                    <Checkbox key={m.key} value={m.key}>{m.label}</Checkbox>
                  ))}
                </Checkbox.Group>
              </div>
            </Space>
            {error && <Alert type="error" message={error} style={{ marginTop: 12 }} />}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="优势对比（最佳/亚军）" size="small">
            <Row gutter={[8, 8]}>
              {computeAdvantage().map((a, i) => (
                <Col span={12} key={i}>
                  <Card size="small" type={a.isGood ? '' : 'inner'} style={{background: a.isGood ? '#f0f7ff' : '#fff'}}>
                    <Statistic title={a.metric} value={a.advantageRatio} precision={2} suffix="倍"
                      valueStyle={{ color: a.isGood ? '#1890ff' : '#52c41a' }} />
                    <Meta description={`${a.bestVariant} vs ${a.runnerUp}`} />
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
        <Col span={24}>
          <Card title="参数对比表（含实验不确定度）" size="small">
            <Table rowKey="variantCode" size="small" columns={columns} dataSource={selectedVariants} pagination={false} scroll={{ x: 1200 }} />
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="雷达图（越靠外越好）" size="small">
            <ResponsiveContainer width="100%" height={380}>
              <RadarChart data={radarData}>
                <PolarGrid />
                <PolarAngleAxis dataKey="metric" tick={{fontSize:11}} />
                <PolarRadiusAxis angle={30} domain={[0, 'auto']} />
                {selectedVariants.map((v, i) => (
                  <Radar key={v.variantCode} name={v.name} dataKey={v.variantCode} stroke={['#1890ff','#52c41a','#faad14'][i%3]} fill={['#1890ff','#52c41a','#faad14'][i%3]} fillOpacity={0.25} />
                ))}
                <Legend />
              </RadarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="射速对比（发/分钟）" size="small">
            <ResponsiveContainer width="100%" height={380}>
              <BarChart data={selectedVariants.map(v => ({ name: v.name, 射速: getMetaVal(v, 'idealFireRate') }))}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Bar dataKey="射速" fill="#1890ff" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default MechanismComparison;
