package database

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/config"
	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

// Models 全部业务模型（AutoMigrate 使用）。
var Models = []interface{}{
	&model.User{},
	&model.Plot{},
	&model.PlantingPlan{},
	&model.HarvestRecord{},
	&model.DiaryEntry{},
	&model.DiaryComment{},
	&model.CommunityPost{},
	&model.CommunityComment{},
	&model.AuditLog{},
}

// Connect 建立 PostgreSQL 连接并完成迁移与种子数据。
func Connect(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	gormWriter := gormlogger.Writer(log.New(os.Stdout, "\r\n", log.LstdFlags))
	gormLogger := gormlogger.New(gormWriter, gormlogger.Config{
		SlowThreshold:             500 * time.Millisecond,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, err
	}
	logger.Info(constants.LogDBConnected, "host", cfg.DBHost, "port", cfg.DBPort, "db", cfg.DBName)

	if err := db.AutoMigrate(Models...); err != nil {
		return nil, err
	}
	logger.Info(constants.LogDBMigrateDone, "tables", len(Models))

	if err := Seed(db, logger); err != nil {
		return nil, err
	}
	return db, nil
}

// Seed 初始化种子数据（仅当用户表为空时执行）。
func Seed(db *gorm.DB, logger *slog.Logger) error {
	var userCount int64
	if err := db.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount > 0 {
		return nil
	}

	seedUsers := []model.User{
		{Username: "admin", Nickname: "平台管理员", Email: "admin@communitygarden.local", Role: string(constants.RoleAdmin), Status: string(constants.UserStatusActive)},
		{Username: "farmer", Nickname: "都市农场主老李", Email: "farmer@communitygarden.local", Role: string(constants.RoleFarmer), Status: string(constants.UserStatusActive)},
		{Username: "citizen", Nickname: "阳台种菜小张", Email: "citizen@communitygarden.local", Role: string(constants.RoleCitizen), Status: string(constants.UserStatusActive)},
	}
	for i := range seedUsers {
		hash, err := util.HashPassword(seedUsers[i].Username + "123")
		if err != nil {
			return err
		}
		seedUsers[i].Password = hash
		if err := db.Create(&seedUsers[i]).Error; err != nil {
			return err
		}
	}

	now := time.Now()
	farmerID := seedUsers[1].ID
	adopterID := farmerID
	seedPlots := []model.Plot{
		{Name: "阳光一区 A01", Code: "P-001", Area: 12.5, SoilType: string(constants.SoilLoam), Sunlight: string(constants.SunlightFull), Latitude: 31.2304, Longitude: 121.4737, Status: string(constants.PlotStatusAvailable), Description: "全日照壤土，适合番茄与辣椒"},
		{Name: "阳光一区 A02", Code: "P-002", Area: 10.0, SoilType: string(constants.SoilBlack), Sunlight: string(constants.SunlightFull), Latitude: 31.2311, Longitude: 121.4742, Status: string(constants.PlotStatusAdopted), AdopterID: &adopterID, Description: "黑土肥力足，老李认养中"},
		{Name: "雨水二区 B01", Code: "P-003", Area: 8.8, SoilType: string(constants.SoilClay), Sunlight: string(constants.SunlightPartial), Latitude: 31.2298, Longitude: 121.4751, Status: string(constants.PlotStatusAvailable), Description: "半日照黏土，适合叶菜与根茎"},
		{Name: "雨水二区 B02", Code: "P-004", Area: 15.2, SoilType: string(constants.SoilSand), Sunlight: string(constants.SunlightFull), Latitude: 31.2289, Longitude: 121.4748, Status: string(constants.PlotStatusHarvested), AdopterID: &adopterID, Description: "沙土排水好，等待释放"},
		{Name: "阳台三区 C01", Code: "P-005", Area: 6.5, SoilType: string(constants.SoilLoam), Sunlight: string(constants.SunlightShade), Latitude: 31.2309, Longitude: 121.4729, Status: string(constants.PlotStatusAvailable), Description: "遮阴区，推荐薄荷与韭菜"},
		{Name: "阳台三区 C02", Code: "P-006", Area: 9.0, SoilType: string(constants.SoilBlack), Sunlight: string(constants.SunlightPartial), Latitude: 31.2301, Longitude: 121.4733, Status: string(constants.PlotStatusAvailable), Description: "半日照黑土，香草专区"},
	}
	for i := range seedPlots {
		if err := db.Create(&seedPlots[i]).Error; err != nil {
			return err
		}
	}

	plantDate := now.AddDate(0, 0, -20)
	harvestDate := util.NextHarvestDate(plantDate, string(constants.CropVegetable))
	plan := model.PlantingPlan{
		PlotID:              seedPlots[1].ID,
		UserID:              farmerID,
		CropName:            "番茄",
		CropType:            string(constants.CropVegetable),
		Season:              string(constants.SeasonSummer),
		Status:              string(constants.PlanStatusGrowing),
		PlantDate:           &plantDate,
		ExpectedHarvestDate: &harvestDate,
		Notes:               "老李的夏季番茄试验田",
	}
	if err := db.Create(&plan).Error; err != nil {
		return err
	}

	seedDiaries := []model.DiaryEntry{
		{PlanID: plan.ID, UserID: farmerID, ActionType: string(constants.DiarySowing), Title: "播种日", Content: "今天完成了番茄播种，用育苗盘催芽，三天后出苗。", LikeCount: 3},
		{PlanID: plan.ID, UserID: farmerID, ActionType: string(constants.DiaryWatering), Title: "日常浇水", Content: "早晚各浇一次水，保持土壤湿润但不积水。", LikeCount: 1},
	}
	for i := range seedDiaries {
		if err := db.Create(&seedDiaries[i]).Error; err != nil {
			return err
		}
	}

	seedPosts := []model.CommunityPost{
		{UserID: farmerID, Title: "番茄立枯病防治经验分享", Content: "发现幼苗茎基部变褐，立刻用多菌灵灌根并减少浇水，一周后恢复。", PostType: string(constants.PostPest), Status: string(constants.PostStatusPublished), LikeCount: 5, CommentCount: 2},
		{UserID: seedUsers[2].ID, Title: "阳台香草拼盘怎么做", Content: "薄荷+罗勒+迷迭香组合，浇水见干见湿，剪枝后长得更旺。", PostType: string(constants.PostExperience), Status: string(constants.PostStatusPublished), LikeCount: 8, CommentCount: 1},
		{UserID: farmerID, Title: "周末集体除草活动召集", Content: "本周六上午 9 点在一区集合，一起除草顺便交流施肥心得，欢迎带工具。", PostType: string(constants.PostActivity), Status: string(constants.PostStatusPublished), LikeCount: 12, CommentCount: 3},
	}
	for i := range seedPosts {
		if err := db.Create(&seedPosts[i]).Error; err != nil {
			return err
		}
	}

	seedComments := []model.CommunityComment{
		{PostID: seedPosts[0].ID, UserID: seedUsers[2].ID, Content: "学习了，多菌灵浓度大概多少？"},
		{PostID: seedPosts[2].ID, UserID: seedUsers[0].ID, Content: "算我一个，带锄头！"},
	}
	for i := range seedComments {
		if err := db.Create(&seedComments[i]).Error; err != nil {
			return err
		}
	}

	logger.Info(constants.LogDBSeedDone, "users", len(seedUsers), "plots", len(seedPlots))
	return nil
}
