package util

import (
	"fmt"
	"time"

	"github.com/communitygarden/server/internal/constants"
)

// formatters.go 同时包含日期、状态文本、类型文本等格式化逻辑（多处耦合点）。
// 新增状态/枚举值必须同步修改此处。

// FormatDate 日期格式化 yyyy-MM-dd。
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTime 日期时间格式化 yyyy-MM-dd HH:mm。
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// RoleText 角色中文文本。
func RoleText(role string) string {
	switch constants.RoleType(role) {
	case constants.RoleAdmin:
		return "管理员"
	case constants.RoleFarmer:
		return "农场主"
	case constants.RoleCitizen:
		return "城市居民"
	default:
		return "未知角色"
	}
}

// PlotStatusText 地块状态中文文本。
func PlotStatusText(s string) string {
	switch constants.PlotStatus(s) {
	case constants.PlotStatusAvailable:
		return "空闲可认养"
	case constants.PlotStatusAdopted:
		return "已认养"
	case constants.PlotStatusHarvested:
		return "待释放"
	default:
		return "未知状态"
	}
}

// PlanStatusText 种植计划状态中文文本。
func PlanStatusText(s string) string {
	switch constants.PlanStatus(s) {
	case constants.PlanStatusPlanned:
		return "已计划"
	case constants.PlanStatusPlanting:
		return "播种中"
	case constants.PlanStatusGrowing:
		return "生长中"
	case constants.PlanStatusHarvesting:
		return "采收中"
	case constants.PlanStatusCompleted:
		return "已完成"
	default:
		return "未知状态"
	}
}

// CropTypeText 作物类型中文文本。
func CropTypeText(t string) string {
	switch constants.CropType(t) {
	case constants.CropVegetable:
		return "蔬菜"
	case constants.CropFruit:
		return "水果"
	case constants.CropHerb:
		return "香草"
	default:
		return "未知类型"
	}
}

// SeasonText 季节中文文本。
func SeasonText(s string) string {
	switch constants.Season(s) {
	case constants.SeasonSpring:
		return "春季"
	case constants.SeasonSummer:
		return "夏季"
	case constants.SeasonAutumn:
		return "秋季"
	case constants.SeasonWinter:
		return "冬季"
	default:
		return "未知季节"
	}
}

// SoilTypeText 土壤类型中文文本。
func SoilTypeText(s string) string {
	switch constants.SoilType(s) {
	case constants.SoilLoam:
		return "壤土"
	case constants.SoilClay:
		return "黏土"
	case constants.SoilSand:
		return "沙土"
	case constants.SoilBlack:
		return "黑土"
	default:
		return "未知土壤"
	}
}

// SunlightText 日照条件中文文本。
func SunlightText(s string) string {
	switch constants.Sunlight(s) {
	case constants.SunlightFull:
		return "全日照"
	case constants.SunlightPartial:
		return "半日照"
	case constants.SunlightShade:
		return "遮阴"
	default:
		return "未知日照"
	}
}

// DiaryActionText 种植日记动作中文文本。
func DiaryActionText(a string) string {
	switch constants.DiaryAction(a) {
	case constants.DiarySowing:
		return "播种"
	case constants.DiaryWatering:
		return "浇水"
	case constants.DiaryFertilizing:
		return "施肥"
	case constants.DiaryPestControl:
		return "除虫"
	case constants.DiaryHarvest:
		return "收成"
	case constants.DiaryOther:
		return "其他"
	default:
		return "未知动作"
	}
}

// HarvestQualityText 收成品质中文文本。
func HarvestQualityText(q string) string {
	switch constants.HarvestQuality(q) {
	case constants.QualityExcellent:
		return "优"
	case constants.QualityGood:
		return "良"
	case constants.QualityFair:
		return "一般"
	default:
		return "未知品质"
	}
}

// PostTypeText 社区帖子类型中文文本。
func PostTypeText(t string) string {
	switch constants.PostType(t) {
	case constants.PostExperience:
		return "种植经验"
	case constants.PostPest:
		return "病虫害防治"
	case constants.PostRecipe:
		return "食谱创意"
	case constants.PostActivity:
		return "线下农耕活动"
	default:
		return "未知类型"
	}
}

// PostStatusText 社区帖子状态中文文本。
func PostStatusText(s string) string {
	switch constants.PostStatus(s) {
	case constants.PostStatusPublished:
		return "已发布"
	case constants.PostStatusRemoved:
		return "已删除"
	default:
		return "未知状态"
	}
}

// WeightKgText 重量文本（kg 保留 2 位小数）。
func WeightKgText(w float64) string {
	return fmt.Sprintf("%.2f kg", w)
}

// NextHarvestDate 根据播种日期与作物类型估算成熟日期（蔬菜 45 天、水果 90 天、香草 35 天）。
func NextHarvestDate(plantDate time.Time, cropType string) time.Time {
	switch constants.CropType(cropType) {
	case constants.CropFruit:
		return plantDate.AddDate(0, 0, 90)
	case constants.CropHerb:
		return plantDate.AddDate(0, 0, 35)
	default:
		return plantDate.AddDate(0, 0, 45)
	}
}

// DaysUntil 距目标日期的天数。
func DaysUntil(target time.Time) int {
	return int(time.Until(target).Hours() / 24)
}
