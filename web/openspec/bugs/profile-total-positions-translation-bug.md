# Profile账户概览持仓数翻译缺失 - Bug修复提案

## 报告信息
- **报告ID**: BUG-2025-12-06-001
- **页面**: https://www.agentrade.xyz/profile
- **语言**: 中文
- **优先级**: P2（用户体验）
- **状态**: 待修复

## 问题描述

### 现象层 - 用户看到的
- 登录后切换中文，进入`帐户概览`卡片
- 持仓数标签显示为 `[Total Positions]`，未被翻译为中文

### 根因分析
- `web/src/pages/UserProfilePage.tsx` 中账户概览的持仓数文案调用 `t('totalPositions', language)`
- 翻译键 `totalPositions` 仅定义在 `profile` 命名空间下 (`translations.ts:533`)，顶层不存在该键
- 翻译函数在找不到键时返回可读占位 `[Total Positions]`，导致中文显示英文占位符

### 复现步骤
1. 登录 https://www.agentrade.xyz/profile
2. 在右上角切换语言为中文
3. 查看“账户概览”卡片右下角的持仓数标签，实际显示 `[Total Positions]`

### 影响范围
- `/profile` 页面中文模式用户
- 账户概览卡片受影响（交易员概览卡片已使用正确的 `profile.totalPositions` 键）

## 修复方案
- 将账户概览统计卡片的文案调用改为 `t('profile.totalPositions', language)`，与翻译文件的键路径保持一致
- 继续复用现有翻译文案 `profile.totalPositions: '总持仓数'`

## 验证方案
- 手动：切换中文后访问 `/profile`，确认持仓数标签显示“总持仓数”且不再出现方括号占位
- 自动：静态检查 `UserProfilePage.tsx` 不再包含 `t('totalPositions'` 这样的顶层调用
