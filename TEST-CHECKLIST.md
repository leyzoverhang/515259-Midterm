# เช็คลิสต์ก่อนอัดวิดีโอ demo — favorite & rating

ทำตามลำดับนี้ทั้งหมด ถ้าข้อไหนไม่ผ่านให้หยุดแก้ก่อน อย่าข้ามไปข้อถัดไป
(อ้างอิงเกณฑ์ "เสร็จ" จากคู่มือสอบ หมวด 09 และกับดัก 6 ข้อ หมวด 05)

## 0. Build + automated tests

```
cd api
go build ./...          # ต้องผ่าน ไม่มี error
go vet ./...             # ต้องสะอาด
go test ./... -count=1   # ทุกเทสต้องเขียว (ต้องเปิด Docker ไว้ก่อน เพราะมี Testcontainers)
```

ถ้า `go test` ฟ้อง compile error ในไฟล์ mock ให้รัน `mockery` ใหม่ก่อน (interface `Repository`/`Service` เปลี่ยนไปแล้ว)

## 1. ยกระบบ + migrate

```
docker compose up -d              # postgres / redis / keycloak
cd api
goose -dir migrations postgres "<วาง POSTGRES_DSN จากไฟล์ .env ตรงนี้>" up
```

เช็คว่า migrate ผ่านครบ (ต้องเห็นตาราง `user_favorites`, `recipe_ratings`, และ `difficulties`/`durations` มี seed data 3+4 แถว)

```
docker exec -it wongnok-postgres psql -U postgres -d test -c "\dt"
docker exec -it wongnok-postgres psql -U postgres -d test -c "select * from difficulties"
```

รันเซิร์ฟเวอร์:

```
go run ./cmd/api
```

เปิด Swagger เช็คว่าเห็น endpoint ใหม่ครบ: `http://localhost:8080/swagger/index.html`
ต้องเห็น `POST /recipes/{id}/favorite`, `DELETE /recipes/{id}/favorite`, `POST /recipes/{id}/rating`,
และ `GET /recipes` ต้องมี query param `favorite` ให้เลือก

## 2. เอา access token จริงมาใช้ยิง (ไม่มี frontend รันอยู่ ต้องทำมือ)

1. เปิดเบราว์เซอร์ไปที่ `http://localhost:8080/api/v1/auth/login`
2. ล็อกอินด้วย user ใน Keycloak realm `pea` (สร้าง user ใน `http://localhost:8081` admin console ก่อนถ้ายังไม่มี — admin/admin)
3. เบราว์เซอร์จะ redirect ไป `http://localhost:3000/auth/callback?ticket=xxxxx` แล้วขึ้น "เชื่อมต่อไม่ได้" เพราะไม่มี frontend รันอยู่ — **ไม่เป็นไร ให้ copy ค่า `ticket` จาก URL ทันที** (มีอายุแค่ 30 วินาที ต้องรีบ)
4. เปิด Postman/curl ยิง exchange ทันที:

```
curl -X POST http://localhost:8080/api/v1/auth/exchange \
  -H "Content-Type: application/json" \
  -d "{\"ticket\":\"<ticket ที่ copy มา>\"}"
```

จะได้ `{ "accessToken": "...", "refreshToken": "...", ... }` — เก็บ `accessToken` ไว้ใช้ทุก endpoint ถัดไป
(access token หมดอายุเร็ว ถ้าเวลาผ่านไปนานให้กลับไปทำข้อ 1-4 ใหม่)

ตั้งตัวแปรไว้ให้สะดวก (bash):
```
TOKEN="<accessToken ที่ได้>"
```

## 3. เตรียมสูตรทดสอบ 1 อัน (ถ้ายังไม่มี)

```
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"Test Recipe","description":"demo","difficultyId":"easy","durationId":"30m","ingredients":[{"description":"test"}],"instructions":[{"description":"test"}]}'
```
คาดหวัง `201 { "id": <ตัวเลข> }` — เก็บ id ไว้เป็น `$RID`

## 4. Checklist ตามเกณฑ์ "เสร็จ" (หมวด 09) — ทำทีละข้อ ติ๊กถูกทุกอัน

- [ ] **401 ไม่มี token** — `curl -i http://localhost:8080/api/v1/recipes` (ไม่ใส่ header) → ต้องได้ `401`
- [ ] **400 id ไม่ใช่ตัวเลข** — `curl -i -X POST http://localhost:8080/api/v1/recipes/abc/favorite -H "Authorization: Bearer $TOKEN"` → `400`
- [ ] **404 สูตรไม่มีจริง** — `curl -i -X POST http://localhost:8080/api/v1/recipes/999999/favorite -H "Authorization: Bearer $TOKEN"` → `404`
- [ ] **Favorite ปกติ** — `curl -i -X POST http://localhost:8080/api/v1/recipes/$RID/favorite -H "Authorization: Bearer $TOKEN"` → `204`
- [ ] **GET /recipes/:id เห็น isFavorite:true** — `curl http://localhost:8080/api/v1/recipes/$RID -H "Authorization: Bearer $TOKEN"` → เช็ค `"isFavorite": true` ใน response
- [ ] **T2/T3: filter favorite** — `curl "http://localhost:8080/api/v1/recipes?favorite=true" -H "Authorization: Bearer $TOKEN"` เห็น `$RID` / `?favorite=false` ไม่เห็น `$RID` / ไม่ใส่ query เห็นทั้งหมด
- [ ] **T1: unfavorite แล้ว favorite ใหม่ ต้องเหลือ 1 แถว** —
  ```
  curl -X DELETE http://localhost:8080/api/v1/recipes/$RID/favorite -H "Authorization: Bearer $TOKEN"   # 204
  curl -X POST   http://localhost:8080/api/v1/recipes/$RID/favorite -H "Authorization: Bearer $TOKEN"   # 204 อีกครั้ง ไม่ error
  docker exec -it wongnok-postgres psql -U postgres -d test -c "select count(*) from user_favorites where recipe_id=$RID"
  ```
  ต้องได้ `count = 1` เท่านั้น (ถ้า 2 = โค้ดใช้ INSERT ธรรมดา ผิด trap T1)
- [ ] **Rating ปกติ + average คำนวณถูก** —
  ```
  curl -i -X POST http://localhost:8080/api/v1/recipes/$RID/rating -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"rating":4}'   # 204
  curl http://localhost:8080/api/v1/recipes/$RID -H "Authorization: Bearer $TOKEN"   # rating.average = 4, rating.total = 1
  ```
- [ ] **T4/T5: ให้คะแนนซ้ำโดยคนเดิม ต้องแทนที่ ไม่ใช่เพิ่มแถว** —
  ```
  curl -X POST http://localhost:8080/api/v1/recipes/$RID/rating -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"rating":2}'
  docker exec -it wongnok-postgres psql -U postgres -d test -c "select count(*) from recipe_ratings where recipe_id=$RID"
  ```
  ต้องได้ `count = 1` และ `rating.average` จาก GET ต้องเป็น `2` (ค่าล่าสุด ไม่ใช่ค่าเฉลี่ยของ 4 กับ 2)
- [ ] **rating นอกช่วง 1-5 → 400** — `-d '{"rating":6}'` และ `-d '{"rating":0}'` → ทั้งคู่ `400`
- [ ] **ไม่เหลือคะแนนเลย average ต้องเป็น 0 ไม่ error (T5)** — ทดสอบกับสูตรใหม่ที่ยังไม่มีใครให้คะแนน → `rating.average: 0`

## 5. ซ้อมอธิบายปากเปล่า

เปิด `EXAM-NOTES.md` อ่านออกเสียงทีละข้อ (Q1–Q8) โดยไม่ดูจอ ถ้าติดข้อไหนให้เปิดโค้ดจริงตรงไฟล์ที่อ้างถึงแล้วอธิบายซ้ำจนคล่อง
ข้อ Q8 (JWT) มีรายละเอียดเพิ่มจากที่แก้ล่าสุด: ตอนนี้ middleware ใช้ `verifier.Verify()` จริง (ตรวจ signature/issuer/expiry) ผ่าน verifier ตัวที่สองที่ `SkipClientIDCheck: true` เพราะ token ที่ frontend ส่งมาเป็น **access token** (aud=account) ไม่ใช่ id token — เตรียมตอบเผื่อกรรมการถามลึกตรงนี้ด้วย

## 6. ก่อนกด push/อัดจริง

```
git status        # ต้องไม่มีอะไรค้าง
git push origin main
```
เข้า GitHub เช็คว่าไฟล์ขึ้นครบและ `.env` ไม่หลุดไปด้วย
