-- 清空表并重置序列（如果需要）
-- TRUNCATE 'table-name' RESTART IDENTITY;--需要+CASCADE！
-- 保持数据但修复序列与数据的一致性（如果使用PostgreSQL）
-- SELECT setval('table-name_id_seq', (SELECT MAX(id) FROM table-name);

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