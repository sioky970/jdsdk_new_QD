package constants

// 通用消息
const (
	MsgSuccess       = "操作成功"
	MsgSystemBusy    = "系统繁忙，请稍后重试"
	MsgParamError    = "请求参数错误"
	MsgUnauthorized  = "未授权或登录已过期"
	MsgForbidden     = "无权执行此操作"
	MsgNotFound      = "未找到相关信息"
	MsgInternalError = "系统错误，请稍后重试"
)

// ========== 认证模块 ==========
const (
	// 注册相关
	MsgRegisterSuccess    = "注册成功，欢迎加入！"
	MsgRegisterParamError = "注册信息填写不完整，请检查用户名和密码"
	MsgUsernameExists     = "该用户名已被注册，请换一个用户名试试"
	MsgRegisterFailed     = "注册失败，请稍后重试或联系管理员"

	// 登录相关
	MsgLoginSuccess       = "登录成功，欢迎回来！"
	MsgLoginParamError    = "请输入用户名和密码"
	MsgLoginFailed        = "用户名或密码不正确，请重新输入"
	MsgLoginSessionFailed = "登录会话创建失败，请重试"

	// 登出相关
	MsgLogoutSuccess = "已安全退出，期待您再次光临！"

	// Token相关
	MsgTokenExpired   = "登录已过期，请重新登录"
	MsgTokenRefreshed = "登录状态已刷新"
	MsgTokenGenFailed = "登录会话创建失败，请重试"
)

// ========== 用户模块 ==========
const (
	// 用户基本操作
	MsgUserNotFound    = "未找到该用户信息"
	MsgUserInfoUpdated = "用户信息已更新"
	MsgUserCreated     = "用户创建成功"
	MsgUserDeleted     = "用户已删除"

	// 密码相关
	MsgPasswordChanged       = "密码修改成功，请使用新密码登录"
	MsgPasswordOldIncorrect  = "当前密码输入错误，请重新输入"
	MsgPasswordWrong         = "当前密码输入错误，请重新输入" // 与MsgPasswordOldIncorrect保持一致
	MsgPasswordTooShort      = "新密码至少需要6个字符，请重新设置"
	MsgPasswordParamError    = "请输入旧密码和新密码"
	MsgPasswordEncryptFailed = "密码加密失败，请稍后重试"

	// API密钥相关
	MsgAPIKeyGenerated      = "API密钥已生成，请妥善保管"
	MsgAPIKeyRefreshed      = "API密钥已刷新，旧密钥已失效"
	MsgAPIKeyGenFailed      = "API密钥生成失败，请重试"
	MsgApiKeyGenerateFailed = "API密钥生成失败，请重试" // 与MsgAPIKeyGenFailed保持一致

	// 京豆操作
	MsgJingdouRecharged    = "京豆充值成功！"
	MsgJingdouAdjusted     = "京豆余额已调整"
	MsgJingdouInsufficient = "京豆余额不足，无法完成扣减"
)

// ========== 任务模块 ==========
const (
	// 任务创建
	MsgTaskCreated             = "任务创建成功！已消耗 %d 京豆"
	MsgTaskCreateParamError    = "任务信息不完整，请检查必填项"
	MsgTaskCreateFailed        = "任务创建失败，请稍后重试"
	MsgTaskTypeInvalid         = "选择的任务类型不存在，请重新选择"
	MsgTaskTypeDisabled        = "该任务类型暂时不可用，请选择其他类型"
	MsgTaskTimeSlotLimit       = "该任务类型仅在 %s 开放创建"
	MsgTaskBalanceInsufficient = "京豆余额不足，当前需要 %d 京豆，您的余额为 %d 京豆"

	// 任务查询
	MsgTaskNotFound = "未找到该任务，可能已被取消"

	// 任务取消
	MsgTaskCancelled    = "任务已取消，%d 京豆已退还"
	MsgTaskCannotCancel = "只有等待执行的任务可以取消"
	MsgTaskNoPermission = "您没有权限操作该任务"

	// 任务更新
	MsgTaskUpdated          = "任务已更新"
	MsgTaskUpdateParamError = "请提供需要更新的任务信息"
	MsgTaskDeleted          = "任务已删除"
	MsgTaskPriorityUpdated  = "任务优先级已更新"

	// 任务类型管理
	MsgTaskTypePreset   = "系统已预设所有任务类型，如需调整请联系管理员"
	MsgTaskTypeUpdated  = "任务类型配置已更新"
	MsgTaskTypeNotFound = "任务类型不存在"

	// 批量操作
	MsgBatchTaskEmpty   = "请至少添加一个任务"
	MsgBatchTaskLimit   = "一次最多可创建100个任务，请分批提交"
	MsgBatchTaskCreated = "成功创建 %d 个任务，共消耗 %d 京豆"
)

// ========== 设备模块 ==========
const (
	MsgDeviceNotFound      = "未找到该设备信息"
	MsgDeviceStatusUpdated = "设备状态已更新"
	MsgDeviceParamError    = "设备信息不完整"
	MsgDeviceNoTask        = "暂无可执行任务"
	MsgDeviceTaskAssigned  = "任务分配成功"
	MsgDeviceTaskFeedback  = "任务执行结果已提交"
	MsgDeviceCleared       = "所有设备数据已清空"
)

// ========== 京豆模块 ==========
const (
	// 京豆日志不需要额外消息，使用通用Success即可
	MsgJingdouBalanceQuery = "" // 查询操作无需提示消息
)

// ========== 设置模块 ==========
const (
	MsgSettingSaved           = "配置已保存"
	MsgSettingUpdated         = "设置已更新"
	MsgSettingInitialized     = "默认配置已初始化"
	MsgSettingParamError      = "配置信息不完整"
	MsgSettingNotFound        = "配置项不存在"
	MsgSettingParseFailed     = "配置数据格式错误"
	MsgSettingSerializeFailed = "配置保存失败，请重试"
)

// ========== API日志模块 ==========
const (
	// API日志查询不需要额外消息
	MsgAPILogQuery = ""
)

// ========== 仪表板模块 ==========
const (
	// 仪表板查询不需要额外消息
	MsgDashboardQuery = ""
)
