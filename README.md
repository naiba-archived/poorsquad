# PoorSquad ![构建状态](https://github.com/naiba/poorsquad/workflows/Build%20Docker%20Image/badge.svg)

:call_me_hand: 穷逼小分队，GitHub 账号管理面板。专为中小型团队、工作室在 GitHub 愉快协作管理雇员使用。

## 界面预览

| ![面板首页.png](https://i.loli.net/2019/11/30/mnpwvNe3j7Es2WC.png) | ![企业主页.png](https://i.loli.net/2019/11/30/2tLDa618KTb4lEo.png) | ![站点登录.png](https://i.loli.net/2019/11/30/2OzkryKDcYXLGq9.png) |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| 面板主页                                                           | 企业主页                                                           | 站点登录                                                           |

## 部署说明

1. 注册一个 [OAuth2 应用](https://github.com/settings/developers)，callback URL 填写 `[你的站点地址]/oauth2/callback`
2. 创建 `data` 文件夹，参考 `data/config.yaml.example` 进行配置
3. 参考 `docker-compose.yaml` 在 Docker 中启动

## 基本功能

- 雇员
  - 超级管理员：第一个登录到系统的人，具有最高（所有）权限
    - 创建仓库
    - 删除仓库
  - 企业：每个雇员可以自由添加企业
    - 企业管理员
      - 管理企业绑定的账号
      - 管理企业团队
      - 绑定团队项目
    - 企业成员：查看企业信息
    - 绑定的 GitHub 账号
    - 项目
      - 外部贡献者：单项目外部贡献者，只能阅读单项目内信息
      - webhook：添加、修改、删除、触发 webhook
      - branch：添加、删除保护分支 ****todo***
    - 小组
      - 小组管理员：管理小组所属项目的
        - Webhook
        - Protect Branch ****todo***
        - Deploy Key ****todo***
        - 项目下成员
      - 小组成员
        - 读取项目的所有信息
        - 触发 webhook

## 版权声明

MIT License
