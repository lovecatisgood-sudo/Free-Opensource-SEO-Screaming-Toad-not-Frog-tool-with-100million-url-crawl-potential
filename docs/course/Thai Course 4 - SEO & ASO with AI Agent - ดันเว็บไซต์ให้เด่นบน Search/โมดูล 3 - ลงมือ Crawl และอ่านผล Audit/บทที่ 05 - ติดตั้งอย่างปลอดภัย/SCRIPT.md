# บทที่ 5 — ติดตั้งอย่างปลอดภัยก่อนเริ่ม Crawl

**เวลาเป้าหมาย:** 4 นาที

## บทพูด

[ภาพบนจอ: Repository ทางการของ `lovecatisgood-sudo`]

ก่อนรันโปรแกรม Open Source ที่โหลดจากอินเทอร์เน็ต อย่าเริ่มจากการ Copy คำสั่งแล้วกด Enter ทันที ให้ตรวจเจ้าของ Repository, README, License, Security Model, Project State และ Commit ล่าสุดก่อน

สำหรับ Toad ให้ใช้ Repository ทางการภายใต้บัญชี `lovecatisgood-sudo` เราสามารถให้ Codex หรือ Claude Code ช่วยตรวจได้ด้วย Prompt นี้

[ภาพบนจอ: Prompt ภาษาไทย]

“Clone Repository ทางการของ SEO Screaming Toad จาก `lovecatisgood-sudo` อ่าน README, Security Model, Development Guide และ Project State ก่อนรันคำสั่ง อธิบาย Go และ Node Version ที่ต้องใช้ ตรวจ Build Script และขออนุญาตก่อนติดตั้ง Dependency ระดับระบบ จากนั้น Build และเปิด Local App แต่ยังห้าม Crawl เว็บไซต์”

จุดสำคัญคือให้ Agent อ่านก่อนลงมือ และแยกสิทธิ์ระหว่างการ Build ในเครื่องกับการส่ง Request ออกไปยังเว็บไซต์

[ภาพบนจอ: `make bootstrap`, `go run ./cmd/seo-auditor`, `127.0.0.1:7331`]

ตัวโปรแกรมปกติเปิดที่ Loopback `127.0.0.1:7331` หมายความว่าออกแบบให้ใช้ในเครื่องเรา ไม่ใช่เอาไปเปิดเป็น Public Server ส่วน Node และ pnpm จำเป็นเมื่อพัฒนา Frontend หรือใช้โหมด JavaScript Rendering

ครั้งแรกควรใช้เว็บที่เราเป็นเจ้าของ ตั้ง Limit เล็ก และเริ่มด้วย Raw Mode ถ้าเจอ Error อย่าแก้ด้วยการปิด TLS, Robots หรือ Network Guard ให้ Agent แสดงคำสั่ง ข้อความผิดพลาด และเอกสารที่เกี่ยวข้อง แล้วแก้จากสาเหตุจริง

การติดตั้งที่ปลอดภัยอาจช้ากว่าการกดรันแบบไม่อ่านเพียงไม่กี่นาที แต่ช่วยลดความเสี่ยงได้มากครับ
