package recipe

import (
	"context"
	"testing"
	"time"

	"wongnok/internal/platform/database"
	"wongnok/internal/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

// เทสชุดนี้ต้องมี Docker รันอยู่ (Rancher Desktop / Docker Desktop) เพราะ
// Testcontainers จะยกคอนเทนเนอร์ Postgres จริงขึ้นมาทดสอบ SQL ตรง ๆ
// ต่างจาก service_test.go ที่ mock repository ไว้ทั้งหมด
//
// หมายเหตุ: ถ้า `go build` บ่นว่า package testcontainers-go / modules/postgres
// หาไม่เจอ ให้รัน (แล้ว go mod tidy อีกที):
//
//   go get github.com/testcontainers/testcontainers-go@latest
//   go get github.com/testcontainers/testcontainers-go/modules/postgres@latest
//
// ถ้าเวอร์ชันที่ได้เก่ากว่า v0.30 ฟังก์ชัน postgres.Run อาจจะชื่อ
// postgres.RunContainer แทน (API เปลี่ยนชื่อระหว่างเวอร์ชัน) แก้ตรงบรรทัด
// ที่เรียกได้เลยถ้าเจอ compile error ตรงนี้

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("wongnok_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = pgContainer.Terminate(context.Background())
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, sqldb, err := database.Open(ctx, dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqldb.Close()
	})

	// AutoMigrate แทนการรัน migrations/*.sql ตรง ๆ — พอสำหรับเทสระดับ
	// repository เพราะสิ่งที่ต้องการคือ schema ที่ตรงกับ struct tag จริง
	// (composite primary key ของ user_favorites/recipe_ratings สำคัญที่สุด
	// เพราะ ON CONFLICT ของกับดัก T1/T4 อาศัย constraint นี้)
	require.NoError(t, db.AutoMigrate(
		&user.User{},
		&Difficulty{},
		&Duration{},
		&Recipe{},
		&RecipeIngredient{},
		&RecipeInstruction{},
		&UserFavorite{},
		&RecipeRating{},
	))

	return db
}

// seedRecipe สร้างข้อมูลอ้างอิงที่ Recipe ต้องการ (difficulty/duration/creator)
// แล้วคืน id ของสูตรที่สร้างไว้ให้เทสต่อได้เลย
func seedRecipe(t *testing.T, db *gorm.DB) int {
	t.Helper()

	// FirstOrCreate กันพัง ถ้าเทสเดียวกันเรียก seedRecipe หลายครั้ง
	// (difficulty/duration เป็นข้อมูลอ้างอิงที่ตั้งใจให้ใช้ร่วมกันได้)
	difficulty := Difficulty{ID: "easy", Name: "Easy"}
	require.NoError(t, db.Where(Difficulty{ID: "easy"}).FirstOrCreate(&difficulty).Error)

	duration := Duration{ID: "short", Name: "Short"}
	require.NoError(t, db.Where(Duration{ID: "short"}).FirstOrCreate(&duration).Error)

	name := "Chef Test"
	creator := user.User{ID: uuid.New(), Email: "chef@example.com", Name: &name, UID: uuid.NewString()}
	require.NoError(t, db.Create(&creator).Error)

	r := Recipe{
		Name:         "Pad Thai",
		Description:  "Test recipe",
		DifficultyID: difficulty.ID,
		DurationID:   duration.ID,
		CreatorID:    creator.ID,
	}
	require.NoError(t, db.Create(&r).Error)

	return r.ID
}

// T1: กดโปรด → ยกเลิก → กดใหม่ ต้องยังเหลือ 1 แถว (ไม่ใช่ duplicate key error)
// และ deleted_at ต้องกลับเป็น NULL
func TestRepository_Favorite_RestoresAfterUnfavorite(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	recipeID := seedRecipe(t, db)

	require.NoError(t, repo.Favorite(ctx, userID, recipeID))
	require.NoError(t, repo.Unfavorite(ctx, userID, recipeID))
	// กดใหม่หลังยกเลิก — ถ้ายังใช้ INSERT ธรรมดาตรงนี้จะ error duplicate key
	require.NoError(t, repo.Favorite(ctx, userID, recipeID))

	var count int64
	require.NoError(t, db.Model(&UserFavorite{}).
		Where("user_id = ? AND recipe_id = ?", userID, recipeID).
		Count(&count).Error)
	require.Equal(t, int64(1), count, "ต้องมีแถวเดียว ไม่ใช่แถวใหม่ซ้อนแถวเก่า")

	favorites, err := repo.FavoriteRecipeIDs(ctx, userID, []int{recipeID})
	require.NoError(t, err)
	require.True(t, favorites[recipeID], "หลังกดใหม่ต้องนับเป็น favorite อีกครั้ง")
}

// T3: filter favorite=true ต้องเห็นเฉพาะของ user ที่ล็อกอินอยู่ ไม่ปนของคนอื่น
func TestRepository_List_FavoriteFilterIsPerUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	owner := uuid.New()
	someoneElse := uuid.New()

	recipeA := seedRecipe(t, db)
	recipeB := seedRecipe(t, db)

	require.NoError(t, repo.Favorite(ctx, owner, recipeA))
	require.NoError(t, repo.Favorite(ctx, someoneElse, recipeB))

	favoriteTrue := true
	recipes, total, err := repo.List(ctx, owner, GetRecipesQuery{
		Pagination: Pagination{Page: 1, Limit: 12},
		Favorite:   &favoriteTrue,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, recipes, 1)
	require.Equal(t, recipeA, recipes[0].ID, "ต้องเห็นเฉพาะสูตรที่ owner กดโปรดเอง")
}

// T4/T5: ให้คะแนนซ้ำ (คนเดิม) ต้องได้ 1 แถว ไม่ใช่ 2 และ average ต้องอัปเดตตามค่าล่าสุด
func TestRepository_Rate_UpsertsAndRecomputesAverage(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	userA := uuid.New()
	userB := uuid.New()
	recipeID := seedRecipe(t, db)

	require.NoError(t, repo.Rate(ctx, userA, recipeID, 5))
	require.NoError(t, repo.Rate(ctx, userB, recipeID, 3))

	var avgAfterTwo float64
	require.NoError(t, db.Model(&Recipe{}).Where("id = ?", recipeID).
		Pluck("average_rating", &avgAfterTwo).Error)
	require.InDelta(t, 4.0, avgAfterTwo, 0.0001, "average ของ 5 กับ 3 ต้องเป็น 4")

	// userA เปลี่ยนใจให้คะแนนใหม่ (ต้อง upsert ไม่ใช่เพิ่มแถว)
	require.NoError(t, repo.Rate(ctx, userA, recipeID, 1))

	var countA int64
	require.NoError(t, db.Model(&RecipeRating{}).
		Where("user_id = ? AND recipe_id = ?", userA, recipeID).
		Count(&countA).Error)
	require.Equal(t, int64(1), countA, "ให้คะแนนซ้ำต้องมีแถวเดียวต่อ user")

	var avgAfterUpdate float64
	require.NoError(t, db.Model(&Recipe{}).Where("id = ?", recipeID).
		Pluck("average_rating", &avgAfterUpdate).Error)
	require.InDelta(t, 2.0, avgAfterUpdate, 0.0001, "average ของ 1 กับ 3 ต้องเป็น 2")

	counts, err := repo.RatingCounts(ctx, []int{recipeID})
	require.NoError(t, err)
	require.Equal(t, int64(2), counts[recipeID], "มี user ให้คะแนน 2 คน ไม่ว่าจะให้กี่ครั้งก็ตาม")
}
