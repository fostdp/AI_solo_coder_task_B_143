import React, { useState, useEffect, useMemo } from 'react'
import {
  Card,
  Button,
  Checkbox,
  Table,
  Tag,
  Spin,
  Alert,
  Space,
  Typography,
  Row,
  Col,
  Statistic,
  Empty
} from 'antd'
import {
  SwapOutlined,
  TrophyOutlined,
  ThunderboltOutlined,
  AimOutlined,
  BoxPlotOutlined
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { variantApi } from '../services/api'

const { Title, Text } = Typography

const MOCK_VARIANTS = [
  {
    code: 'zhuge',
    name: '诸葛弩',
    dynasty: '三国',
    description: '诸葛亮发明的连发弩，箭匣容量大，可持续射击'
  },
  {
    code: 'san-gong',
    name: '三弓弩',
    dynasty: '宋代',
    description: '大型床弩，三张弓合力，射程极远，需多人操作'
  },
  {
    code: 'bi-zhang',
    name: '臂张弩',
    dynasty: '战国',
    description: '单人臂力上弦，轻便灵活，便于携带'
  }
]

const MOCK_COMPARE_RESULT = {
  variants: MOCK_VARIANTS.map(v => ({
    ...v,
    armLength: v.code === 'san-gong' ? 220 : v.code === 'zhuge' ? 85 : 65,
    stringTension: v.code === 'san-gong' ? 1800 : v.code === 'zhuge' ? 450 : 280,
    magazineCapacity: v.code === 'zhuge' ? 10 : v.code === 'san-gong' ? 1 : 1,
    maxRange: v.code === 'san-gong' ? 1500 : v.code === 'zhuge' ? 260 : 180,
    effectiveRange: v.code === 'san-gong' ? 800 : v.code === 'zhuge' ? 150 : 100,
    theoreticalFireRate: v.code === 'zhuge' ? 12 : v.code === 'san-gong' ? 0.5 : 2,
    reloadTime: v.code === 'zhuge' ? 8 : v.code === 'san-gong' ? 120 : 15,
    accuracyScore: v.code === 'bi-zhang' ? 88 : v.code === 'zhuge' ? 72 : 65,
    sustainabilityScore: v.code === 'zhuge' ? 85 : v.code === 'bi-zhang' ? 60 : 30
  })),
  parameterComparison: [
    { parameter: '弩臂长度', unit: 'cm', values: { 'zhuge': 85, 'san-gong': 220, 'bi-zhang': 65 }, bestCode: 'bi-zhang' },
    { parameter: '弓弦张力', unit: 'N', values: { 'zhuge': 450, 'san-gong': 1800, 'bi-zhang': 280 }, bestCode: 'san-gong' },
    { parameter: '箭匣容量', unit: '发', values: { 'zhuge': 10, 'san-gong': 1, 'bi-zhang': 1 }, bestCode: 'zhuge' },
    { parameter: '最大射程', unit: 'm', values: { 'zhuge': 260, 'san-gong': 1500, 'bi-zhang': 180 }, bestCode: 'san-gong' },
    { parameter: '有效射程', unit: 'm', values: { 'zhuge': 150, 'san-gong': 800, 'bi-zhang': 100 }, bestCode: 'san-gong' },
    { parameter: '理论射速', unit: '发/分', values: { 'zhuge': 12, 'san-gong': 0.5, 'bi-zhang': 2 }, bestCode: 'zhuge' },
    { parameter: '装弹时间', unit: 's', values: { 'zhuge': 8, 'san-gong': 120, 'bi-zhang': 15 }, bestCode: 'zhuge' },
    { parameter: '精度评分', unit: '分', values: { 'zhuge': 72, 'san-gong': 65, 'bi-zhang': 88 }, bestCode: 'bi-zhang' }
  ],
  radarData: {
    dimensions: ['射速', '张力', '射程', '弹容', '射速可持续性', '精度'],
    series: [
      { code: 'zhuge', name: '诸葛弩', values: [90, 45, 35, 95, 88, 70] },
      { code: 'san-gong', name: '三弓弩', values: [10, 98, 99, 10, 20, 60] },
      { code: 'bi-zhang', name: '臂张弩', values: [35, 30, 28, 10, 55, 92] }
    ]
  },
  advantages: [
    { category: '射速之王', code: 'zhuge', name: '诸葛弩', value: 12, unit: '发/分' },
    { category: '射程之王', code: 'san-gong', name: '三弓弩', value: 1500, unit: 'm' },
    { category: '弹容之王', code: 'zhuge', name: '诸葛弩', value: 10, unit: '发' },
    { category: '便携之王', code: 'bi-zhang', name: '臂张弩', value: 2.5, unit: 'kg' }
  ]
}

const CrossbowVariantCompare = () => {
  const [variants, setVariants] = useState(MOCK_VARIANTS)
  const [selectedCodes, setSelectedCodes] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)
  const [initialLoading, setInitialLoading] = useState(true)

  useEffect(() => {
    const loadVariants = async () => {
      try {
        const data = await variantApi.getVariants()
        if (data && data.length > 0) {
          setVariants(data)
        }
      } catch (err) {
        console.warn('使用默认弩型数据:', err.message)
      } finally {
        setInitialLoading(false)
      }
    }
    loadVariants()
  }, [])

  const handleCheckboxChange = (code, checked) => {
    setSelectedCodes(prev =>
      checked ? [...prev, code] : prev.filter(c => c !== code)
    )
  }

  const handleCompare = async () => {
    if (selectedCodes.length < 2) return
    setLoading(true)
    setError(null)
    try {
      const data = await variantApi.compareVariants(selectedCodes)
      setResult(data)
    } catch (err) {
      console.warn('使用模拟对比数据:', err.message)
      const filteredResult = {
        ...MOCK_COMPARE_RESULT,
        variants: MOCK_COMPARE_RESULT.variants.filter(v => selectedCodes.includes(v.code)),
        parameterComparison: MOCK_COMPARE_RESULT.parameterComparison.map(p => ({
          ...p,
          values: Object.fromEntries(
            Object.entries(p.values).filter(([k]) => selectedCodes.includes(k))
          )
        })),
        radarData: {
          ...MOCK_COMPARE_RESULT.radarData,
          series: MOCK_COMPARE_RESULT.radarData.series.filter(s => selectedCodes.includes(s.code))
        },
        advantages: MOCK_COMPARE_RESULT.advantages.filter(a => selectedCodes.includes(a.code))
      }
      setResult(filteredResult)
    } finally {
      setLoading(false)
    }
  }

  const selectedVariants = variants.filter(v => selectedCodes.includes(v.code))

  const tableColumns = useMemo(() => {
    const cols = [
      {
        title: '参数名',
        dataIndex: 'parameter',
        key: 'parameter',
        fixed: 'left',
        width: 140,
        render: (text, record) => (
          <Space>
            <Text strong>{text}</Text>
            <Text type="secondary">({record.unit})</Text>
          </Space>
        )
      }
    ]
    selectedVariants.forEach(v => {
      cols.push({
        title: (
          <Space direction="vertical" size={0} style={{ alignItems: 'center' }}>
            <Text strong style={{ color: '#B8860B' }}>{v.name}</Text>
            <Text type="secondary" style={{ fontSize: 12 }}>{v.dynasty}</Text>
          </Space>
        ),
        dataIndex: `value_${v.code}`,
        key: v.code,
        width: 140,
        align: 'center',
        render: (_, record) => {
          const val = record.values[v.code]
          const isBest = record.bestCode === v.code
          return (
            <div style={{
              background: isBest ? 'rgba(46, 139, 87, 0.15)' : 'transparent',
              borderRadius: 6,
              padding: '4px 8px'
            }}>
              <Text strong style={{ color: isBest ? '#2E8B57' : '#d4c4a8', fontSize: 15 }}>
                {val ?? '-'}
              </Text>
              {isBest && <Tag color="green" icon={<TrophyOutlined />} style={{ marginLeft: 4, fontSize: 11 }}>最优</Tag>}
            </div>
          )
        }
      })
    })
    return cols
  }, [selectedVariants])

  const tableData = useMemo(() => {
    if (!result) return []
    return result.parameterComparison.map((p, idx) => ({
      key: idx,
      ...p
    }))
  }, [result])

  const radarOption = useMemo(() => {
    if (!result) return {}
    const colorMap = { 'zhuge': '#B8860B', 'san-gong': '#CD5C5C', 'bi-zhang': '#4682B4' }
    return {
      tooltip: {},
      legend: {
        data: result.radarData.series.map(s => s.name),
        textStyle: { color: '#d4c4a8' },
        bottom: 0
      },
      radar: {
        indicator: result.radarData.dimensions.map(d => ({
          name: d,
          max: 100
        })),
        splitArea: {
          areaStyle: {
            color: ['rgba(139, 69, 19, 0.1)', 'rgba(139, 69, 19, 0.2)']
          }
        },
        axisLine: { lineStyle: { color: '#8B4513' } },
        splitLine: { lineStyle: { color: '#8B4513', opacity: 0.5 } },
        name: { textStyle: { color: '#d4c4a8' } }
      },
      series: [{
        type: 'radar',
        data: result.radarData.series.map(s => ({
          value: s.values,
          name: s.name,
          lineStyle: { color: colorMap[s.code] || '#B8860B', width: 2 },
          itemStyle: { color: colorMap[s.code] || '#B8860B' },
          areaStyle: {
            color: colorMap[s.code] || '#B8860B',
            opacity: 0.15
          }
        }))
      }]
    }
  }, [result])

  const advantageIconMap = {
    '射速之王': <ThunderboltOutlined />,
    '射程之王': <AimOutlined />,
    '弹容之王': <BoxPlotOutlined />,
    '便携之王': <SwapOutlined />
  }

  if (initialLoading) {
    return (
      <div style={{ padding: 24, display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ padding: 24, overflow: 'auto', height: '100%' }}>
      <Card
        style={{
          background: 'linear-gradient(135deg, #2a1f18 0%, #1a1410 100%)',
          border: '1px solid #8B4513',
          marginBottom: 16
        }}
        bodyStyle={{ padding: 24 }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Title level={3} style={{ margin: 0, color: '#B8860B', fontFamily: "'Ma Shan Zheng', cursive" }}>
              <SwapOutlined style={{ marginRight: 8 }} />
              弩型机构对比分析
            </Title>
            <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
              选择至少2种弩型进行多维度参数对比与雷达图分析
            </Text>
          </div>

          <div>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              选择弩型（至少2个）：
            </Text>
            <Row gutter={[16, 16]}>
              {variants.map(v => (
                <Col xs={24} sm={8} key={v.code}>
                  <Card
                    size="small"
                    hoverable
                    style={{
                      background: selectedCodes.includes(v.code)
                        ? 'rgba(184, 134, 11, 0.15)'
                        : 'rgba(42, 31, 24, 0.6)',
                      border: selectedCodes.includes(v.code)
                        ? '2px solid #B8860B'
                        : '1px solid #5a4a3a',
                      cursor: 'pointer',
                      transition: 'all 0.3s'
                    }}
                    onClick={() => handleCheckboxChange(v.code, !selectedCodes.includes(v.code))}
                  >
                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                      <Space>
                        <Checkbox
                          checked={selectedCodes.includes(v.code)}
                          onChange={(e) => handleCheckboxChange(v.code, e.target.checked)}
                        />
                        <Text strong style={{ color: selectedCodes.includes(v.code) ? '#B8860B' : '#d4c4a8' }}>
                          {v.name}
                        </Text>
                        <Tag color="#8B4513">{v.dynasty}</Tag>
                      </Space>
                      <Text type="secondary" style={{ fontSize: 12, lineHeight: 1.5 }}>
                        {v.description}
                      </Text>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </div>

          <Space>
            <Button
              type="primary"
              size="large"
              icon={<SwapOutlined />}
              onClick={handleCompare}
              disabled={selectedCodes.length < 2 || loading}
              loading={loading}
              style={{
                background: 'linear-gradient(135deg, #B8860B 0%, #8B4513 100%)',
                border: 'none',
                minWidth: 160
              }}
            >
              加载对比（{selectedCodes.length}/3）
            </Button>
            <Text type="secondary">
              {selectedCodes.length < 2 && `还需选择 ${2 - selectedCodes.length} 个弩型`}
            </Text>
          </Space>
        </Space>
      </Card>

      {error && (
        <Alert
          type="error"
          message={error}
          showIcon
          closable
          style={{ marginBottom: 16 }}
          onClose={() => setError(null)}
        />
      )}

      {loading && (
        <Card style={{ background: '#2a1f18', border: '1px solid #8B4513' }}>
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size="large" />
            <div style={{ marginTop: 16, color: '#a89888' }}>正在分析对比数据...</div>
          </div>
        </Card>
      )}

      {result && !loading && (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Card
            title={
              <Space>
                <TrophyOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>参数对比表格</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
          >
            <Table
              dataSource={tableData}
              columns={tableColumns}
              pagination={false}
              size="middle"
              scroll={{ x: 'max-content' }}
              rowKey="key"
              style={{
                background: 'transparent'
              }}
            />
          </Card>

          <Row gutter={16}>
            <Col xs={24} lg={14}>
              <Card
                title={
                  <Space>
                    <BoxPlotOutlined style={{ color: '#B8860B' }} />
                    <span style={{ color: '#d4c4a8' }}>综合性能雷达图</span>
                  </Space>
                }
                style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
                headStyle={{ borderBottom: '1px solid #8B4513' }}
              >
                <div style={{ height: 400 }}>
                  <ReactECharts
                    option={radarOption}
                    style={{ height: '100%', width: '100%' }}
                    notMerge
                    lazyUpdate
                    opts={{ renderer: 'canvas' }}
                  />
                </div>
              </Card>
            </Col>

            <Col xs={24} lg={10}>
              <Card
                title={
                  <Space>
                    <TrophyOutlined style={{ color: '#B8860B' }} />
                    <span style={{ color: '#d4c4a8' }}>各项优势标注</span>
                  </Space>
                }
                style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
                headStyle={{ borderBottom: '1px solid #8B4513' }}
              >
                {result.advantages && result.advantages.length > 0 ? (
                  <Row gutter={[12, 12]}>
                    {result.advantages.map((adv, idx) => (
                      <Col span={12} key={idx}>
                        <Card
                          size="small"
                          style={{
                            background: 'linear-gradient(135deg, rgba(184, 134, 11, 0.2) 0%, rgba(139, 69, 19, 0.2) 100%)',
                            border: '1px solid #B8860B',
                            textAlign: 'center',
                            height: '100%'
                          }}
                        >
                          <div style={{ fontSize: 28, color: '#B8860B', marginBottom: 8 }}>
                            {advantageIconMap[adv.category] || <TrophyOutlined />}
                          </div>
                          <Tag
                            color="#B8860B"
                            style={{
                              fontSize: 13,
                              padding: '2px 12px',
                              marginBottom: 8,
                              borderRadius: 10
                            }}
                          >
                            {adv.category}
                          </Tag>
                          <div style={{ color: '#d4c4a8', marginBottom: 4, fontWeight: 500 }}>
                            {adv.name}
                          </div>
                          <Statistic
                            value={adv.value}
                            suffix={adv.unit}
                            valueStyle={{ color: '#2E8B57', fontSize: 20 }}
                          />
                        </Card>
                      </Col>
                    ))}
                  </Row>
                ) : (
                  <Empty description="暂无优势数据" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                )}
              </Card>
            </Col>
          </Row>
        </Space>
      )}
    </div>
  )
}

export default CrossbowVariantCompare
