🚀 URL Shortener Service

Một dịch vụ rút gọn URL được xây dựng bằng Golang, sử dụng Gin, PostgreSQL, Docker, và Clean Architecture.
Mục tiêu của project là tạo ra một hệ thống có khả năng tạo các mã URL ngắn ổn định, chống trùng lặp, có khả năng chịu tải cao và dễ mở rộng.

🧩 1. Mô tả bài toán

Bài toán yêu cầu xây dựng một hệ thống rút gọn URL giống như Bitly hoặc TinyURL:

Người dùng gửi original URL ⇒ Hệ thống trả về short code.

Khi người dùng truy cập short URL ⇒ Redirect về original URL.

Có cơ chế:

Không tạo duplicate code.

Nếu cùng một URL đã tồn tại trước đó thì phải xử lý logic tùy yêu cầu.

Giới hạn expired time và tự cập nhật khi hết hạn.

Hỗ trợ concurrency (nhiều request cùng tạo một URL).

Tôi hiểu bài toán theo hướng: phải xây dựng một API đáng tin cậy, tốc độ cao, bảo toàn tính nhất quán trong database, và có thể sẵn sàng mở rộng khi traffic tăng.

🛠️ 2. Cách chạy project
🔧 Yêu cầu hệ thống

Go 1.22+

Docker & Docker Compose

PostgreSQL (local hoặc container)

Makefile đã cài đặt

📦 Step-by-step
1️⃣ Clone project
git clone https://github.com/datmaithanh/URL-Shortener-Service
cd URL-Shortener-Service

2️⃣ Tạo file môi trường
cp app.env.example app.env


Điền:

DB_SOURCE=postgresql://user:password@localhost:5432/shortener?sslmode=disable
SERVER_ADDRESS=0.0.0.0:8080

3️⃣ Chạy PostgreSQL bằng Docker
make postgres

4️⃣ Chạy migration
make migrateup

5️⃣ Chạy server
make server

6️⃣ Test API

Tạo URL rút gọn

POST /api/shorten
{
  "originalUrl": "https://google.com"
}


Redirect

GET /:shortCode

🏗️ 3. Thiết kế & Quyết định kỹ thuật
🗄️ Tại sao chọn PostgreSQL?
Lý do	Giải thích
Tính nhất quán mạnh	Short code phải unique tuyệt đối, PostgreSQL hỗ trợ transaction và unique index rất tốt
Hỗ trợ JSONB	Linh hoạt mở rộng
Sequence & chức năng random tốt	Dễ implement thuật toán generate mã
Mở rộng dễ	Khi scale sang read-replica vẫn ổn

PostgreSQL phù hợp cho một dịch vụ CRUD nhỏ nhưng yêu cầu tính toàn vẹn cao.

🔌 Thiết kế API
POST   /api/shorten       → tạo short url
GET    /:shortCode        → redirect
GET    /api/urls/:id      → lấy thông tin url


Lý do:

RESTful, đơn giản, dễ dùng.

Hướng tới production-ready.

URL redirect nên dùng GET trực tiếp cho tốc độ.

🔑 Thuật toán generate short code

Dùng base62 encoding (0-9, a-z, A-Z):

Generate random 8 ký tự bằng charset Base62
⇒ 62⁸ ≈ 218 nghìn tỷ combination ⇒ gần như không trùng.

Check trong DB:

Nếu tồn tại ⇒ generate lại.

Lưu với unique index để đảm bảo không bao giờ trùng.

Pseudo:

for {
    code := RandomBase62(8)
    if !ExistsInDB(code) {
        return code
    }
}

⚔️ Xử lý conflict / duplicate
TH1: Duplicate short code

→ DB UNIQUE constraint xử lý, app retry tự động.

TH2: Cùng một URL gửi 2 lần

Kiểm tra bằng original_url UNIQUE.

Nếu URL đã rút gọn tồn tại → Trả về short code cũ.

Nếu expired → Update expired_at + extend time.

Giải pháp đảm bảo idempotent cho API.

⚖️ 4. Trade-offs (Lựa chọn kỹ thuật)
🧩 Vì sao chọn random code thay vì incremental ID?

Chọn random Base62 vì:

Incremental ID	Random Base62
Dễ đoán, dễ bị spam	Khó đoán, an toàn
Short URL ngắn	Cũng ngắn
Cần encode ID	Không cần
Dễ trùng khi scale	Không bao giờ trùng

➡️ Chọn random Base62 vì bảo mật & dễ scaling.

🧩 Vì sao chọn Gin?

Nhẹ, nhanh, đơn giản.

Ecosystem mạnh.

Phù hợp microservice.

🧩 Vì sao chọn Clean Architecture?

Tách biệt layers: Handler → Service → Repository

Dễ mở rộng (chuyển DB, cache, queue…)

Dễ viết unit test.

🧨 5. Challenges đã gặp & Cách giải quyết
1️⃣ Vấn đề concurrency: 2 request cùng lúc tạo URL giống nhau

Giải pháp:

DB unique index để đảm bảo consistency.

Retry logic khi conflict.

2️⃣ Vấn đề expired URL

Giải pháp:

Khi check URL nếu đã expired → renew thời gian → update DB.

3️⃣ Vấn đề migration lộ thông tin DB

Giải pháp:

Không gắn DB source vào Makefile.

Đọc env trong runtime hoặc dùng dotenv cho migrate script.

4️⃣ Clone repo & build bị lỗi version hoặc dependency

Giải pháp:

Tạo Makefile chuẩn.

Dùng Docker cho môi trường thống nhất.

🚧 6. Limitations & Improvements
🔸 Hiện tại còn thiếu gì?

Chưa có caching (Redis) để tăng tốc redirect.

Chưa có rate-limit.

Chưa có authentication.

Chưa có system design cho high scale.

🔸 Nếu có thêm thời gian, tôi sẽ làm:

Thêm Redis caching cho short-url → tăng tốc gấp 10 lần.

Thêm Prometheus metrics.

Ghi log bằng Zap hoặc Zerolog.

Tối ưu random generator (sử dụng crypto/rand).

Viết 100% unit test cho service & handler.

🔸 Production-ready cần:

Docker + CI/CD.

Auto migration.

Health check endpoint.

Retry logic mạnh hơn.

Circuit breaker nếu DB quá tải.

SSO hoặc token-based admin API.

📄 7. Kiến trúc tổng quan
/cmd
/internal
    /api        → HTTP handlers
    /service    → business logic
    /repository → db queries
/pkg
    /utils      → helper, random generator
/db
    /migration


=> Dễ mở rộng và clean.
