-- 清空表并重置序列（如果需要）
-- TRUNCATE 'table-name' RESTART IDENTITY;--需要+CASCADE！

-- 插入常见环境1 室内正常环境
INSERT INTO environments (
    name, summary, temperature, humidity, pm25, noise, light, 
    wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at
) VALUES (
     '室内普通环境', '室内正常环境，温度适宜，光线较好，通风良好', 3, 3, 5, 4, 2, 5, 5, 5, 3, 5, 5, 5, 4, NOW(), NOW()
);
-- 插入常见环境2 室外晴朗天气
INSERT INTO environments (
     name, summary, temperature, humidity, pm25, noise, light, 
    wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at
) VALUES (
     '室外晴朗天气', 
    '室外正常环境，温度适宜，光线强烈，通风非常好', 
    3, 3, 5, 4, 2, 
    5, 5, 5, 3, 5, 5, 5, 4, 
    NOW(), NOW()
);
-- 插入常见环境3 普通家庭
INSERT INTO environments (
   name, summary, temperature, humidity, pm25, noise, light, 
    wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at
) VALUES (
    '普通家庭', 
    '普通的家庭环境，温度适宜，光线良好，通风良好', 
    3, 3, 5, 4, 2, 
    5, 5, 5, 3, 5, 5, 5, 4, 
    NOW(), NOW()
);
-- 插入常见环境4 车辆维修车间
INSERT INTO environments (
     name, summary, temperature, humidity, pm25, noise, light, 
    wind, flow, rain, pressure, smoke, dust, odor, visibility, created_at, updated_at
) VALUES (
    '车辆维修车间', 
    '一般的车辆维修车间，光线足，通风好，但有一些机械噪音，轻微尾气等', 
    2, 3, 4, 2, 2, 
    5, 5, 5, 3, 3, 3, 4, 4, 
    NOW(), NOW()
);

-- 插入默认的常见场所安全隐患数据
-- 这些隐患作为示例，帮助用户理解场所隐患的概念（物理状态或环境问题）

INSERT INTO hazards (id, uuid, user_id, name, nickname, keywords, description, source, severity, category, created_at) VALUES
(1, 'hazard-001-damaged-guardrail', 1, '护栏破损', '防护栏损坏', '护栏,破损,防护,栏杆', '作业场所的安全护栏出现破损、松动或缺失，无法提供有效的防护作用，存在人员意外跌落的隐患。', '设施老化缺乏维护', 4, 2, NOW()),

(2, 'hazard-002-high-temp-source', 1, '高温热源', '热源隐患', '高温,热源,热表面,防护', '作业场所存在高温设备、管道或表面，缺乏适当的防护罩或警示标识，人员意外接触可能造成烫伤。', '设备防护不当', 4, 2, NOW()),

(3, 'hazard-003-exposed-wire', 1, '电线裸露', '电线隐患', '电线,裸露,绝缘,电气', '作业场所的电线绝缘层破损或老化，导致电线裸露，存在人员意外接触导致触电事故的隐患。', '电线老化缺乏维护', 5, 1, NOW());

-- 插入三个默认风险示例
INSERT INTO risks (id, uuid, user_id, name, nickname, keywords, description, source, severity, created_at) VALUES 
(1, 'risk-001-uuid', 1, '高空坠落', '坠高险', '高空,坠落,安全带', '在≥2米无护栏平台作业时存在坠落风险，可能导致重伤或死亡', '环境', 5, NOW()),
(2, 'risk-002-uuid', 1, '高温烫伤', '烫伤险', '高温,烫伤,防护', '接触高温设备或介质时可能造成皮肤烫伤，温度≥60℃时触发', '设备', 4, NOW()),
(3, 'risk-003-uuid', 1, '触电风险', '电击险', '触电,电击,绝缘', '接触带电设备或线路时可能发生电击事故，电压≥36V时存在风险', '设备', 5, NOW());

-- 插入默认通用技能
-- 通用硬技能
INSERT INTO skills (id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, category, level, created_at) VALUES
(1, 'skill-001-self-care', 1, '生活自理', '基础生活技能', '包括穿衣、吃饭、洗澡、个人卫生等基本生活技能，是成年人应具备的基础能力', 1, 1, 2, 1, NOW()),
(2, 'skill-002-walking', 1, '行走', '步行技能', '正常的步行能力，包括平地行走、上下楼梯、保持平衡等基本移动技能', 2, 1, 2, 1, NOW()),
(3, 'skill-003-basic-communication', 1, '基础沟通', '日常交流', '基本的语言表达和理解能力，能够进行日常对话和信息交换', 1, 2, 2, 1, NOW());

-- 通用软技能
INSERT INTO skills (id, uuid, user_id, name, nickname, description, strength_level, difficulty_level, category, level, created_at) VALUES
(4, 'skill-004-smile', 1, '微笑', '情绪表达', '通过面部表情表达友善和积极情绪，是基本的社交和情绪管理技能', 1, 1, 1, 1, NOW()),
(5, 'skill-005-patience', 1, '耐心', '情绪控制', '在面对困难或等待时保持冷静和坚持的能力，是重要的情绪管理技能', 1, 3, 1, 1, NOW()),
(6, 'skill-006-empathy', 1, '共情', '理解他人', '理解和感受他人情感的能力，是良好人际关系的基础技能', 1, 3, 1, 1, NOW());

-- 插入默认通用法力
-- 理性法力
INSERT INTO magics (id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, category, level, created_at) VALUES
(1, 'magic-001-observation', 1, '观察', '观察力', '通过感官收集和分析环境信息的能力，是认知和学习的基础', 2, 2, 1, 1, NOW()),
(2, 'magic-002-memory', 1, '记忆', '记忆力', '存储和回忆信息的能力，包括短期记忆和长期记忆', 3, 2, 1, 1, NOW()),
(3, 'magic-003-logic', 1, '逻辑思维', '推理能力', '运用逻辑规则进行推理和分析的思维能力', 4, 4, 1, 1, NOW());

-- 感性法力
INSERT INTO magics (id, uuid, user_id, name, nickname, description, intelligence_level, difficulty_level, category, level, created_at) VALUES
(4, 'magic-004-intuition', 1, '直觉', '第六感', '不经过逻辑推理而直接感知和判断的能力', 3, 4, 2, 1, NOW()),
(5, 'magic-005-creativity', 1, '创造力', '创新思维', '产生新颖想法和解决方案的能力，是艺术和创新的基础', 4, 5, 2, 1, NOW()),
(6, 'magic-006-aesthetic', 1, '审美', '美感', '感受和欣赏美的能力，包括对艺术、自然等的审美判断', 3, 3, 2, 1, NOW());

-- 保持数据但修复序列与数据的一致性（如果使用PostgreSQL）
SELECT setval('hazards_id_seq', (SELECT MAX(id) FROM hazards));
SELECT setval('risks_id_seq', (SELECT MAX(id) FROM risks));
SELECT setval('skills_id_seq', (SELECT MAX(id) FROM skills));
SELECT setval('magics_id_seq', (SELECT MAX(id) FROM magics));