# บทที่ 10 — เชื่อม MCP และสั่ง Agent ทำ Audit จากหลักฐาน

**เวลาเป้าหมาย:** 5 นาที

## บทพูด

[ภาพบนจอ: Codex หรือ Claude Code ใน Workspace ว่าง]

เราจะใช้ AI Agent ช่วยติดตั้งแบบมีขั้นตอน เปิด Workspace ใหม่แล้วสั่งว่า

[ภาพบนจอ: Prompt ภาษาไทย]

“ดึง Repository ทางการของ SEO Screaming Toad จากบัญชี `lovecatisgood-sudo` อ่าน README, MCP, Security Model และ Project State ก่อน Build สร้าง `seo-auditor` กับ `seo-auditor-mcp` ในโฟลเดอร์ `bin` เปิด App ที่ `127.0.0.1:7331` แล้วแสดง MCP Configuration ที่ใช้ Absolute Path โดยยังห้าม Crawl จนกว่าผมจะอนุมัติ Scope”

โปรแกรม `seo-auditor` เป็นตัวหลักที่ดูแล Dashboard, Crawler, API และฐานข้อมูล ส่วน `seo-auditor-mcp` เป็น Adapter ที่ AI Client เปิดขึ้นมา เราจึงต้องเปิดโปรแกรมหลักก่อน

ตำแหน่งไฟล์ MCP Configuration อาจเปลี่ยนตาม Version ของ Codex หรือ Claude Code อย่าจำ Path จากคลิปเก่า ให้ Agent ตรวจเอกสารปัจจุบันของ Client ที่ติดตั้งอยู่

[ภาพบนจอ: `health_get` แสดง `api_connected: true`]

เมื่อเชื่อมแล้ว เริ่มด้วย `health_get` ถ้าพร้อมจริงต้องเห็น `api_connected: true` จากนั้นจึงเตรียมคำสั่ง Crawl เช่น

“เว็บไซต์นี้ผมมีสิทธิ์ตรวจ ใช้ Raw Mode จำกัดหนึ่งพัน URL ไม่รวม Subdomain และตัด `/account`, `/cart`, `/checkout` กับ Internal Search ออก ขอให้ Preview Scope และอธิบายรายการที่ถูกตัดก่อน ห้าม Start จนกว่าผมจะยืนยัน”

หลังยืนยัน Agent จะเริ่ม Crawl หนึ่งครั้งด้วย Idempotency Key แล้วตรวจ Status จนถึง Terminal State เมื่อจบให้สั่งต่อว่า

“จัดลำดับ Error และ Warning ที่น่าจะเกิดจาก Template สำหรับทุกข้อให้แสดง Rule ID, จำนวนหน้าที่ได้รับผล, URL ตัวอย่าง, Evidence, Limitation, Root Cause ที่คาด และวิธีตรวจหลังแก้ แยกข้อเท็จจริงออกจากข้อสันนิษฐาน และยังไม่ต้องแก้ Code”

[ภาพบนจอ: Summary → Finding → Evidence → Recommendation]

อย่าขอข้อมูลทั้งหมดแบบไม่จำกัด ให้ Agent ใช้ Pagination แล้วเลือกตัวแทนมาวิเคราะห์ เป้าหมายของ MCP ไม่ใช่ให้ AI พูดแทน Dashboard แต่ให้มันพาเราไปถึงหลักฐานที่ใช้ตัดสินใจได้เร็วขึ้นครับ
