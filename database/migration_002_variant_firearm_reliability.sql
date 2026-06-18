-- ============================================================
-- 古代连弩动力学仿真系统 - 迁移脚本 002
-- 弩型注册表 / 现代轻武器对比 / 箭匣可靠性分析 / 虚拟射击会话
-- ============================================================

-- 弩型注册表（古代连弩分类）
CREATE TABLE IF NOT EXISTS crossbow_variants (
    code VARCHAR(32) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    dynasty VARCHAR(64),
    era_year INT,
    description TEXT,
    draw_weight_n NUMERIC(10,2),
    max_range_m NUMERIC(10,2),
    effective_range_m NUMERIC(10,2),
    ideal_fire_rate NUMERIC(6,2),
    magazine_size INT,
    reload_time_sec NUMERIC(6,2),
    accuracy_score NUMERIC(4,1),
    mechanism_params JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 现代轻武器对比参考表（只读，用于跨时代对比展示）
CREATE TABLE IF NOT EXISTS modern_firearms (
    code VARCHAR(32) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    origin VARCHAR(64),
    intro_year INT,
    type VARCHAR(32),
    cyclic_rate_rpm INT,
    effective_rpm INT,
    magazine_size INT,
    caliber_mm NUMERIC(6,2),
    effective_range_m INT,
    muzzle_velocity_mps NUMERIC(8,2),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 箭匣可靠性分析结果表（保存历史分析）
CREATE TABLE IF NOT EXISTS magazine_reliability_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crossbow_variant_code VARCHAR(32) NOT NULL REFERENCES crossbow_variants(code),
    sim_shots INT NOT NULL,
    sim_time_sec NUMERIC(12,2) NOT NULL,
    jam_probability_per_shot NUMERIC(12,8),
    mtbf_shots NUMERIC(12,2),
    mtbf_hours NUMERIC(12,4),
    failure_mode_distribution JSONB,
    reliability_curve JSONB,
    confidence_interval JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mr_analyses_variant ON magazine_reliability_analyses(crossbow_variant_code);

-- 虚拟射击会话表（记录用户体验数据）
CREATE TABLE IF NOT EXISTS virtual_shoot_sessions (
    session_id UUID PRIMARY KEY,
    crossbow_variant_code VARCHAR(32) NOT NULL REFERENCES crossbow_variants(code),
    user_id VARCHAR(64) DEFAULT 'anonymous',
    shots_fired INT DEFAULT 0,
    jam_count INT DEFAULT 0,
    reload_count INT DEFAULT 0,
    elapsed_sec NUMERIC(12,2) DEFAULT 0,
    average_rpm NUMERIC(8,2) DEFAULT 0,
    max_instant_rpm NUMERIC(8,2) DEFAULT 0,
    final_string_fatigue NUMERIC(6,4) DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    shot_timestamps TIMESTAMPTZ[] DEFAULT ARRAY[]::TIMESTAMPTZ[]
);
CREATE INDEX IF NOT EXISTS idx_vss_variant ON virtual_shoot_sessions(crossbow_variant_code);
CREATE INDEX IF NOT EXISTS idx_vss_user ON virtual_shoot_sessions(user_id);

-- 预填充弩型
INSERT INTO crossbow_variants (code, name, dynasty, era_year, description, draw_weight_n, max_range_m, effective_range_m, ideal_fire_rate, magazine_size, reload_time_sec, accuracy_score) VALUES
('zhuge', '诸葛连弩', '三国·蜀', 234, '诸葛亮改进的连弩，也称元戎弩，铜弩机，箭匣内置十矢，可连发', 950, 250, 80, 10.0, 10, 5.5, 7.2)
ON CONFLICT (code) DO NOTHING;
INSERT INTO crossbow_variants (code, name, dynasty, era_year, description, draw_weight_n, max_range_m, effective_range_m, ideal_fire_rate, magazine_size, reload_time_sec, accuracy_score) VALUES
('san-gong', '三弓弩', '北宋', 1040, '床子弩的一种，三张弓合力，需绞车张弦，专攻坚城重甲', 3500, 1500, 300, 1.5, 1, 45.0, 8.5)
ON CONFLICT (code) DO NOTHING;
INSERT INTO crossbow_variants (code, name, dynasty, era_year, description, draw_weight_n, max_range_m, effective_range_m, ideal_fire_rate, magazine_size, reload_time_sec, accuracy_score) VALUES
('bi-zhang', '臂张弩', '战国-秦', -300, '单兵臂力张弦的标准弩，结构简单，秦军步兵标配', 1500, 450, 150, 4.0, 1, 12.0, 8.0)
ON CONFLICT (code) DO NOTHING;

-- 预填充现代步枪参考
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('m1-garand', 'M1 加兰德', '美国', 1936, '半自动步枪', 50, 45, 8, 7.62, 457, 865, '二战美军制式，8发漏夹')
ON CONFLICT (code) DO NOTHING;
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('ak-47', 'AK-47', '苏联', 1947, '突击步枪', 600, 100, 30, 7.62, 350, 715, '卡拉什尼科夫，世界产量最高')
ON CONFLICT (code) DO NOTHING;
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('m16a1', 'M16A1', '美国', 1967, '突击步枪', 800, 50, 30, 5.56, 500, 990, '小口径高速弹先驱')
ON CONFLICT (code) DO NOTHING;
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('mp5', 'HK MP5', '德国', 1966, '冲锋枪', 800, 120, 30, 9.0, 100, 400, '反恐特种部队标配')
ON CONFLICT (code) DO NOTHING;
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('m249-saw', 'M249 SAW', '美国', 1984, '班用机枪', 900, 200, 200, 5.56, 800, 915, '弹链/弹匣双供弹，火力压制')
ON CONFLICT (code) DO NOTHING;
INSERT INTO modern_firearms (code, name, origin, intro_year, type, cyclic_rate_rpm, effective_rpm, magazine_size, caliber_mm, effective_range_m, muzzle_velocity_mps, notes) VALUES
('desert-eagle', '沙漠之鹰', '以色列/美国', 1983, '大口径手枪', 0, 30, 7, 12.7, 50, 470, '大威力半自动手枪，非制式')
ON CONFLICT (code) DO NOTHING;
