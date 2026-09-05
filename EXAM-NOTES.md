# คำตอบซ้อมก่อนอัดวิดีโอ — Favorite & Rating

เอกสารนี้ตอบคำถามซ้อม 8 ข้อจากคู่มือสอบ (หมวด 10) ตามโค้ดจริงที่อยู่ในโปรเจกต์ตอนนี้
ใช้พูดตามได้เลย หรืออ่านทำความเข้าใจแล้วอธิบายด้วยคำพูดตัวเองตอนอัดวิดีโอ — กรรมการอยากได้แบบหลังมากกว่า

---

## Q1. ทำไมกดโปรดซ้ำหลังยกเลิกถึงต้อง restore ไม่ใช่ insert ใหม่

`user_favorites` มี primary key เป็น `(user_id, recipe_id)` **ไม่รวม `deleted_at`** ดังนั้นถ้ายกเลิกโปรด (soft delete ตั้ง `deleted_at`) แล้วกดโปรดใหม่ด้วย `INSERT` ธรรมดา จะชนกับแถวเดิมที่ยังอยู่ในตาราง (แค่ถูก soft-delete) เกิด `duplicate key value violates unique constraint` ทันที

วิธีแก้ในโค้ด (`repository.go` → `Favorite`) คือใช้ `clause.OnConflict` ให้ Postgres ตรวจ `(user_id, recipe_id)` ชนไหม ถ้าชนให้ `UPDATE` แทน `INSERT` โดยตั้ง `deleted_at = nil` (คืนชีพ) และ `updated_at = time.Now()` ด้วยตัวเอง เพราะ path `ON CONFLICT DO UPDATE` ข้าม hook อัตโนมัติของ GORM ไป

ทำไมไม่ทำแบบ "เช็คก่อนว่ามีไหม แล้วค่อยตัดสินใจ insert หรือ update": เพราะมี race condition — สอง request พร้อมกัน (เช่น ผู้ใช้กดรัว ๆ หรือ 2 tab) อ่านพร้อมกันแล้วเห็น "ไม่มีแถว" ทั้งคู่ ก็จะพยายาม INSERT พร้อมกันแล้วชนกันอยู่ดี `ON CONFLICT` ที่ให้ database จัดการให้ในคำสั่งเดียวจึงปลอดภัยกว่า

## Q2. `*bool` กับ `bool` ต่างกันยังไงในบริบทนี้ ถ้าใช้ `bool` จะเกิดอะไรขึ้น

query parameter `favorite` ต้องแยกได้ 3 สถานะ: ไม่ส่งมาเลย (เอาทุกสูตร) / `true` (เฉพาะที่กด) / `false` (เฉพาะที่ยังไม่กด) `*bool` ทำได้เพราะ `nil` แทน "ไม่ส่งมา" ได้ ส่วน `&true`/`&false` แทนอีกสองสถานะ

ถ้าใช้ `bool` ธรรมดา ค่า default ของ Go คือ `false` เสมอ ทำให้แยกไม่ออกระหว่าง "ผู้ใช้ตั้งใจกรอง `favorite=false`" กับ "ผู้ใช้ไม่ได้ส่ง query นี้มาเลย" ผลคือถ้าไม่ส่ง query มา ระบบจะเข้าใจผิดว่าต้องกรองเอาเฉพาะสูตรที่ยังไม่ถูกกดโปรด ซึ่งผิดจากที่ควรจะเป็น (ควรได้ทุกสูตร)

โค้ดจริง: `dto.go` → `GetRecipesQuery.Favorite *bool` และเช็คด้วย `if query.Favorite != nil` ใน repository

## Q3. ทำไมกรอง favorite ด้วย subquery ไม่ใช้ JOIN

ถ้า `JOIN` ตาราง `recipes` กับ `user_favorites` ตรง ๆ แล้วสูตรหนึ่งมีคนกดโปรดได้หลายคน แถวผลลัพธ์จะซ้ำ (1 แถวต่อ 1 คนที่กด) ต้องตาม `DISTINCT` เพิ่ม แล้วที่แย่กว่าคือ `COUNT` ที่ใช้คำนวณ `total` สำหรับ pagination จะนับผิด (นับซ้ำตามจำนวนคนกด ไม่ใช่จำนวนสูตร)

การใช้ subquery (`SELECT recipe_id FROM user_favorites WHERE user_id = ?`) แล้วเอา `recipes.id IN (...)` หรือ `NOT IN (...)` ทำให้แต่ละสูตรปรากฏแค่ 1 แถวเสมอ ไม่ต้องมี `DISTINCT` และ `Count()` ก่อนหน้านั้นก็ถูกต้อง

จุดที่ต้องระวังคือ **ต้องมี `WHERE user_id = ?` ในซับคิวรีด้วย** ไม่งั้นจะกลายเป็นกรองด้วยรายการโปรดของผู้ใช้ทุกคนรวมกัน ไม่ใช่ของคนที่ล็อกอินอยู่คนเดียว (โค้ดจริงอยู่ใน `repository.go` → `List`)

## Q4. ทำไมการคำนวณ average ต้องอยู่ใน transaction เดียวกับการบันทึกคะแนน

`Rate` ทำ 3 ขั้นตอน: (1) upsert คะแนนของ user คนนี้ (2) คำนวณ `AVG` ใหม่จากทุกคะแนนของสูตรนั้น (3) เขียน `average_rating` กลับที่ตาราง `recipes` ถ้าไม่ห่อด้วย transaction เดียวกัน แล้วมี 2 คนให้คะแนนพร้อมกัน จะเกิดเคสที่คนที่ 2 คำนวณ `AVG` เสร็จ (ยังไม่รวมคะแนนคนแรกที่ commit ไม่ทัน) แล้วเขียนทับค่าที่ถูกต้องกว่าของคนแรกทิ้งไป ทำให้ `average_rating` เพี้ยนถาวรและไม่มีใคร trigger ให้คำนวณใหม่อีก

โค้ดจริงอยู่ใน `repository.go` → `Rate` ห่อทั้ง 3 ขั้นตอนด้วย `repo.db.Transaction(func(tx *gorm.DB) error { ... })`

## Q5. `COALESCE` มีไว้ทำไม ไม่ใส่แล้วพังตรงไหน

`AVG()` ของ SQL เมื่อไม่มีแถวให้คำนวณเลยจะคืนค่า `NULL` ไม่ใช่ `0` ถ้า `Scan` ค่า `NULL` ลงตัวแปร Go ชนิด `float64` ตรง ๆ จะเกิด error ("converting NULL to float64 is unsupported" หรือคล้ายกัน) `COALESCE(AVG(score), 0)` บอก Postgres ว่าถ้าผลลัพธ์เป็น `NULL` ให้ใช้ `0` แทน ทำให้ `Scan` ปลอดภัยเสมอ (โค้ดอยู่ใน `repository.go` → `Rate` ขั้นตอนที่ 2)

## Q6. mock กับ Testcontainers ต่างกันยังไง ทำไมต้องมีทั้งคู่

mock (ใน `service_test.go`) ปลอมชั้น `Repository` ทั้งหมด ใช้พิสูจน์ "กฎธุรกิจ" ของชั้น service เช่น "ถ้าสูตรไม่มีอยู่จริง ต้องไม่เรียก repository เลย" — เทสเร็วเพราะไม่แตะ database จริง แต่พิสูจน์ไม่ได้ว่า SQL ที่เขียนถูกต้อง

Testcontainers ยกฐานข้อมูล Postgres จริงขึ้นมาในเทส ใช้พิสูจน์ว่า "SQL ที่เขียนทำงานถูกกับ database จริง" เช่น เคส T1 (กดโปรด→ยกเลิก→กดใหม่ ต้องเหลือ 1 แถว) เทสนี้ mock ไม่มีทางจับได้เพราะมันไม่ได้รัน SQL จริง มันแค่บันทึกว่า "ถูกเรียก" เท่านั้น

สรุป: mock ตอบคำถาม "logic ถูกไหม" ส่วน container ตอบคำถาม "query ถูกไหม" ต้องมีทั้งคู่เพราะพิสูจน์คนละเรื่อง

## Q7. ทำไม `Repository` เป็น interface ถ้าเป็น struct ตรง ๆ จะเกิดอะไรขึ้น

`service.go` ประกาศ `type Repository interface {...}` (ไม่ใช่ import struct จาก `repository.go` ตรง ๆ) เพราะ "ผู้ใช้เป็นคนกำหนดว่าต้องการอะไร" — service รู้แค่ว่ามันต้องการเมธอดหน้าตาแบบไหน ไม่ต้องรู้ว่าข้างหลังเป็น Postgres, mock, หรือ database อื่น

ถ้าใช้ struct ตรง ๆ (ผูกกับ `*gorm.DB` เต็ม ๆ) การเทสชั้น service จะทำไม่ได้เลยถ้าไม่มี database จริงเสมอ เพราะ compile time ก็บังคับให้ต้องส่ง `*repository` ตัวจริงเข้ามา ปลอมเป็น mock ไม่ได้ — เทส service จะช้าและเปราะ (ต้องพึ่ง DB ทุกเทส) และเทสไม่ได้แยก concern ระหว่าง "logic" กับ "SQL" อีกต่อไป

## Q8. เช็คสิทธิ์ที่ไหน ถ้าไม่มี token จะเกิดอะไรขึ้นและใครเป็นคนตอบ 401

เช็คที่ `middleware.JWT` (`internal/middleware/jwt.go`) ซึ่งผูกกับทุก route ใน group `/recipes` (`main.go`: `recipeGroup.Use(middleware.JWT(...))`) ถ้า header ไม่มี `Authorization: Bearer ...` middleware จะ `AbortWithStatusJSON(401)` ทันทีและ **ไม่ปล่อยให้ request ไปถึง handler เลย**

แต่ในโค้ดของ handler เอง (`Favorite`, `Unfavorite`, `Rate`, `GetRecipe`) ก็มีการเช็คซ้ำอีกชั้นด้วย `reqctx.UserID(ctx.Request.Context())` — เผื่อกรณีที่ route ถูกเรียกโดยไม่ผ่าน middleware จริง (เช่นตอนเทส handler แบบ httptest ที่ข้าม router) จุดนี้เป็น defense-in-depth ไม่ใช่ความซ้ำซ้อนที่ไม่มีประโยชน์

---

## ของแถม — 2 จุดที่แก้ให้ระหว่างช่วยทำ (เผื่อกรรมการถาม)

**บั๊ก pagination**: `repository.go` → `List` ตอนแรกเขียน `db.Order((query.Page-1)*query.Limit)` ซึ่งผิด เพราะ `.Order()` ใช้สำหรับ `ORDER BY` (รับ string) ไม่ใช่ `.Offset()` ที่ใช้ข้ามแถวสำหรับ pagination (รับ int) ผลคือ pagination พังเงียบ ๆ (หน้า 2, 3 จะได้ผลลัพธ์เหมือนหน้า 1 หรือ error) แก้เป็น `.Offset((query.Page-1)*query.Limit).Limit(query.Limit)`

**isFavorite / rating.total เคย hardcode**: `dto.go` → `NewRecipeResponse` เดิมใส่ `IsFavorite: false` และ `Total: 0` ตายตัวเสมอ (มีคอมเมนต์ "ใส่ค่า default ไปก่อน") แก้โดยเพิ่ม `RecipeView` (struct ที่ห่อ `Recipe` กับ `IsFavorite`/`RatingTotal`) แล้วให้ service คำนวณค่าจริงด้วย query แบบ batch 2 คำสั่ง (`FavoriteRecipeIDs`, `RatingCounts`) แทนที่จะ query ทีละสูตร (กัน N+1) — เลือกวิธีนี้เพราะ SQL เดี่ยวที่รวม subquery ทุกอย่างจะซับซ้อนขึ้นมากและอธิบายยากกว่าตอนสอบ
