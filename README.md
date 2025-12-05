# URL Shortener Service

## Mô tả bài toán

### Bài toán cần giải quyết
Xây dựng một hệ thống rút gọn URL cho phép người dùng:
- Chuyển đổi URL dài thành mã ngắn (short code) dễ chia sẻ
- Truy cập URL gốc thông qua mã ngắn
- Theo dõi số lượng truy cập vào mỗi link

### Hiểu về bài toán
Đây là bài toán thiết kế hệ thống phân tán điển hình, yêu cầu cân bằng giữa:
- **Performance**: Phải redirect nhanh
- **Scalability**: Xử lý được hàng triệu URL
- **Uniqueness**: Đảm bảo mỗi short code là duy nhất
- **Collision Handling**: Xử lý trường hợp trùng lặp mã

### Các yêu cầu chức năng
1. **Shorten URL**: `POST /urls` - Tạo mã ngắn từ URL dài
2. **Redirect**: `GET /:code` - Chuyển hướng đến URL gốc
3. **Get an url**: `GET /urls/:id` - Xem chi tiết một url
4. **Get list urls**: `GET /urls?page_id=1&page_size=5` - Xem chi tiết danh sách url  

---

## 🚀 Cách chạy project
### Link demo
https://shortservice.datmt.id.vn/
Khi test với link demo: Ban đầu có thể load chậm vì là server free của render nên sẽ chậm lần đầu.

### Prerequisites
```bash
go v1.25
make
sqlc v1.30.0
migrate 4.19
```

### Bước 1: Clone repository
```bash
git clone https://github.com/datmaithanh/URL-Shortener-Service
cd URL-Shortener-Service
```

### Bước 2: Cài đặt dependencies
```bash
go mod tidy
```

### Bước 3: Setup Database

**Dùng Neon (Khuyến nghị)**
```bash
https://console.neon.tech/
```
<img width="1906" height="934" alt="image" src="https://github.com/user-attachments/assets/dd88ac25-8762-4c28-928b-d56dc58a446f" />
<img width="738" height="591" alt="image" src="https://github.com/user-attachments/assets/c57299ba-882a-4f10-afe6-afbb76936267" />


### Bước 4: Config môi trường
Tạo file `.env.prod` từ template:
```bash
Sửa lại database source từ source ở bước 3
```
Sửa lại các giá trị trong `config/config.go`

### Bước 5: Setup migration
```bash
make createschema
```
Sau đó copy database schema vào file up vừa được tạo.
<img width="738" height="447" alt="image" src="https://github.com/user-attachments/assets/fad876cb-6b67-4e93-a32a-673c8271d0ad" />
```sql
SET TIME ZONE 'Asia/Ho_Chi_Minh';

ALTER DATABASE urlshortsevice SET timezone TO 'Asia/Ho_Chi_Minh';

CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE,
    short_url TEXT UNIQUE, 
    original_url TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    clicks bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT (now()),
    expires_at timestamptz NOT NULL
);

CREATE INDEX ON "urls" ("code");

CREATE INDEX ON "urls" ("original_url");
make createschema
```

### Bước 6: Migration
```bash
make migrateup
```
### Bước 7: SQLC
```bash
make sqlc
```

### Bước 8: Run app
```bash
make run
```

---

## Thiết kế & Quyết định kỹ thuật

### 1. Database: PostgreSQL

**Lý do chọn PostgreSQL:**
- **ACID compliance**: Đảm bảo tính toàn vẹn dữ liệu khi có concurrent requests
- **Index performance**: B-tree index trên `short_code` cho lookup O(log n)


### 2. Schema Design

```sql
SET TIME ZONE 'Asia/Ho_Chi_Minh';

ALTER DATABASE urlshortsevice SET timezone TO 'Asia/Ho_Chi_Minh';

CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE,
    short_url TEXT UNIQUE, 
    original_url TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL DEFAULT '',
    clicks bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT (now()),
    expires_at timestamptz NOT NULL
);

CREATE INDEX ON "urls" ("code");

CREATE INDEX ON "urls" ("original_url");
```

**Quyết định:**
- `code` là code được tạo ra từ thuật toán
- `short_url` là url đã được rút gọn và sẵn sàng phục vụ
- `original_url` là url gốc
- `clicks` là số lượng click vào short link

### 3. Thuật toán Generate Short Code

**Approach: Base62 Encoding + Counter**

```golang
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}
	result := ""
	for num > 0 {
		remainder := num % 62
		result = string(base62Chars[remainder]) + result
		num = num / 62
	}
	return result
}
```

**Tại sao Base62?**
- 62^7 = 3.5 trillion combinations → Đủ scale
- URL-safe characters (không cần encode)
- Deterministic từ database ID

**Alternative approach đã xem xét:**
| Approach | Pros | Cons | Kết luận |
|----------|------|------|----------|
| Random + Retry | Đơn giản | Collision tăng theo thời gian | ❌ Không scale |
| UUID | Unique 100% | Quá dài (36 chars) | ❌ Không ngắn |
| Hash (MD5) | Deterministic | Collision possible, dài | ❌ Cần shorten hash |
| **Base62 + ID** | Unique, ngắn, fast | Cần auto-increment ID | ✅ **Chọn** |

### 4. Xử lý Transaction/Duplicate


```golang
func (store *SQLStore) CreateUrlTx(ctx context.Context, arg CreateUrlTxParams) (CreateUrlTxResult, error) {
	var result CreateUrlTxResult

	err := store.execTx(ctx, func(q *Queries) error {

		var err error
		existingUrl, err := q.GetUrlByOriginalUrl(ctx, arg.OriginalUrl)
		if err == nil {
			result.Url = existingUrl
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
		
		result.Url, err = q.CreateUrl(ctx, arg.CreateUrlParams)
		if err != nil {
			return err
		}

		result.Url, err = arg.AfterCreate(q, &result.Url)
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}
```
```golang
existingUrl, err := server.store.GetUrlByOriginalUrl(ctx, req.OriginalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			payload := db.CreateUrlTxParams{
				CreateUrlParams: db.CreateUrlParams{
					OriginalUrl: req.OriginalURL,
					Title:       req.Title,
					ExpiresAt:   time.Now().Add(config.URL_EXPIRE_DURATION),
				},
				AfterCreate: func(q *db.Queries, url *db.Url) (db.Url, error) {
					code := utils.EncodeBase62(url.ID)
					shortUrl := fmt.Sprintf("%s/%s", config.DOMAIN_NAME, code)

					result, err := q.UpdateCodeUrl(ctx, db.UpdateCodeUrlParams{
						ID:       url.ID,
						Code:     sql.NullString{String: code, Valid: true},
						ShortUrl: sql.NullString{String: shortUrl, Valid: true},
					})
					if err != nil {
						return db.Url{}, err
					}
					return result, nil
				},
			}

			url, err := server.store.CreateUrlTx(ctx, payload)
			fmt.Println("DEBUG URL: ", url)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			response := urlsResponse{
				Code:        url.Url.Code.String,
				ShortURL:    url.Url.ShortUrl.String,
				OriginalURL: url.Url.OriginalUrl,
				Title:       url.Url.Title,
				Clicks:      url.Url.Clicks,
				CreatedAt:   url.Url.CreatedAt,
				ExpiresAt:   url.Url.ExpiresAt,
			}
			ctx.JSON(http.StatusOK, response)
			return
		}
	}
```


### 5. API Design

**RESTful với pragmatic choices:**

```
`POST /urls` - Tạo mã ngắn từ URL dài
`GET /:code` - Chuyển hướng đến URL gốc
`GET /urls/:id` - Xem chi tiết một url`
`GET /urls?page_id=1&page_size=5` - Xem chi tiết danh sách url 
```

---

## ⚖️ Trade-offs

### 1. Database Layer: SQLC vs Alternatives

**Chọn:** SQLC
**Thay vì:** GORM, sqlx, ent, raw database/sql

**Lý do:**
```
✅ Type safety: Generate Go code từ SQL → compile-time errors
✅ Performance: Không có reflection overhead như ORM
✅ SQL-first: Viết SQL thuần → full control query optimization
```
### 2. golang-migrate (Database Migrations)
```
// GORM: Dễ nhưng nguy hiểm
db.AutoMigrate(&URL{}) 
// ❌ Không có down migration
// ❌ Không review được changes
// ❌ Production data loss risk
```
```
// golang-migrate: Explicit
// ✅ Peer review migrations
// ✅ Test rollback
// ✅ Audit trail
```

### 3. Gin Framework
```
Gin:    Simple, fast, popular (67k stars)
Chi:    Thuần Go idioms, stdlib-like, nhẹ hơn
Echo:   Nhiều features hơn (WebSocket, SSE)
```
Ưu điểm nhược điểm:
```
Trade-off đã chấp nhận:

Gin không follow Go idioms 100% (context riêng thay vì stdlib)
Middleware signature khác với net/http standard
Locked vào Gin ecosystem

Lý do vẫn chọn:

Community lớn → nhiều resources, plugins
Performance proven ở production scale
Onboarding developers dễ hơn

```

### 4. Base62 Encoding
Đã chọn Base62:
```
Zero collision: Database sequence đảm bảo unique
Performance: 1 query, không retry
Simple: Không cần random generator, collision check
Scalable: 3.5 trillion combinations với 7 chars
```
##  Challenges & Solutions

### Vấn đề: 2 request 1 time

**Vấn đề:**  
Khi gửi 2 request của 1 lúc với nội dung gióng nhau thì có 2 record giống nhau.

**Giải quyết:**
```
Sử dụng transaction khi đang tạo url mới thì check tồn tại.
Nếu đúng thì mới được tạo hoàn toàn, nếu không đúng thì cho rollback.
```

**Học được:**
```
Xử lí khi record trùng lặp
Kĩ năng debug được nâng cao
```
## 🚧 Limitations & Improvements

### Limitations hiện tại

1. **No JWT**
   - Nó đang là public
   
2. **No Caching Layer**
   - Mọi request đều hit database → Latency cao cho popular links
   
3. **No Rate Limiting**
   - Dễ bị abuse, DDoS

### Nếu có thêm 1 tuần:

**Week Plan:**

**Day 1-2: Caching Layer**
```
- Cache-aside pattern
- Expected: P99 latency giảm nhiều lần
```

**Day 3-5: Security**
```
- Rate limiting: redis-rate-limiter (100 req/IP/hour)
- Paseto/jwt
```
##  Performance Benchmarks

### Điều làm tốt:
- Database schema đơn giản nhưng effective
- Base62 algorithm elegant và scalable
- Xử lý errors đầy đủ
- Xử lý transaction (có thể mở rộng)

### Điều có thể làm tốt hơn:
- Nên thêm caching từ đầu (premature optimization nhưng đáng)
- Auth cho app với Paseto
### Học được:
- Database constraints > Application logic cho uniqueness
- Trade-offs luôn tồn tại, quan trọng là document chúng

---

## Test api
### Create a short url
<img width="1102" height="680" alt="image" src="https://github.com/user-attachments/assets/9e532be6-2035-415a-9367-91ed1b952efc" />

### Redirect a short url to ori url
<img width="728" height="46" alt="image" src="https://github.com/user-attachments/assets/8ab67d69-ad56-4a3a-8d8d-38b8a2a6ad7e" />

### Get a url
<img width="1100" height="658" alt="image" src="https://github.com/user-attachments/assets/c61d87ac-454e-4098-baf0-0f546076547d" />

### Get list url
<img width="1111" height="875" alt="image" src="https://github.com/user-attachments/assets/b12ee613-5bd5-481f-ba8d-8a68bb77fcb3" />


## 👤 Author

**Mai Thanh Dat**
- Email: datmt07@gmail.com
- GitHub: https://github.com/datmaithanh

---

**Last Updated:** December 2025
