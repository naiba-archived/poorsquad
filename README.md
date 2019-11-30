# PoorSquad ![构建状态](https://github.com/naiba/poorsquad/workflows/Build%20Docker%20Image/badge.svg)

:call_me_hand: 穷逼小分队，GitHub 账号管理面板。专为中小型团队、工作室在 GitHub 愉快协作管理雇员使用。

## 界面预览

| ![面板首页.png](https://i.loli.net/2019/11/30/mnpwvNe3j7Es2WC.png) | ![企业主页.png](https://i.loli.net/2019/11/30/2tLDa618KTb4lEo.png) | ![站点登录.png](https://i.loli.net/2019/11/30/2OzkryKDcYXLGq9.png) |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| 面板主页                                                           | 企业主页                                                           | 站点登录                                                           |

## 基本功能

- 雇员
  - 超级管理员：第一个登录到系统的人，具有最高（所有）权限
  - 企业：每个雇员可以自由添加企业
    - 企业管理员
      - 管理企业绑定的账号
      - 管理企业团队
      - 绑定团队项目
    - 企业成员：查看企业信息
    - 绑定的 GitHub 账号
    - 项目
      - 外部贡献者：单项目外部贡献者，只能阅读单项目内信息
      - branch：保护分支、删除分支 ****todo***
      - webhook：添加修改删除触发 webhook ****todo***
    - 小组
      - 小组管理员：管理小组所属项目的
        - Webhook ****todo***
        - Protect Branch ****todo***
        - Deploy Key ****todo***
        - 项目下成员
      - 小组成员
        - 读取项目的所有信息
        - 触发 webhook ****todo***

## 版权声明

MIT License
