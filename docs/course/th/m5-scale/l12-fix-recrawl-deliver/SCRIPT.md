# บทที่ 12 — แก้ด้วย Coding Agent แล้ว Crawl ซ้ำให้เห็นผล

**เวลาเป้าหมาย:** 5 นาที

## บทพูด

[ภาพบนจอ: Audit → จัด Priority → แก้ Code → Test → Deploy → Recrawl → Compare]

Report ยังไม่สร้างผลลัพธ์ให้เว็บไซต์ จนกว่าเราจะเปลี่ยน Finding ที่เชื่อถือได้ให้เป็นการแก้ไข และพิสูจน์ว่าการแก้นั้นไม่สร้างปัญหาใหม่

เริ่มจากกลุ่มเล็กที่มี Evidence ชัดและน่าจะมี Root Cause ร่วมกัน เช่น Title Generator, Canonical Component, Navigation, Sitemap Builder หรือ Hreflang Helper การแก้ Template หนึ่งจุดอาจช่วยหลายร้อยหน้า

ส่งข้อมูลให้ Coding Agent แบบมีขอบเขตว่า

[ภาพบนจอ: Prompt ภาษาไทย]

“ตรวจ Finding จาก Toad ชุดนี้เทียบกับ Source Code แยกข้อเท็จจริงออกจากสมมติฐาน หา Root Cause ร่วม เสนอการแก้ที่เล็กและปลอดภัยที่สุด รักษา Accessibility, Localization, Security และพฤติกรรมเดิม เพิ่ม Test สำหรับ SEO Behavior ที่ต้องการ และยังไม่ Deploy จนกว่าผมจะตรวจ Diff”

สำหรับเว็บ JavaScript ให้ตรวจทั้ง Raw HTML และ Rendered DOM อย่าดูแค่ Browser Inspector เพราะ Search Crawler บางแบบอาจเห็นเนื้อหาคนละช่วงเวลา

เมื่อแก้แล้วให้รัน Test และ Production Build ตรวจ Diff ก่อน Deploy หลังขึ้น Production ต้องเปิด URL ตัวแทนดูจริง แล้วสร้าง Crawl ใหม่ด้วย Configuration ที่เทียบกับ Baseline ได้

[ภาพบนจอ: Compare แสดง Added, Removed, Changed, New Issue และ Fixed Issue]

อย่าเชื่อเพียงว่า Warning ลดลง เปิด Evidence ตัวแทนตรวจอีกครั้ง Canonical ที่แก้ผิดอาจทำให้ทุกหน้าชี้ไปหน้าแรก การเอา noindex ออกอาจเปิดหน้าส่วนตัว และ Sitemap ใหม่อาจเผลอใส่ Redirect

รายงานส่งมอบควรมี Scope, Configuration, Terminal State, Limitation, Finding ที่จัดลำดับแล้ว, สิ่งที่แก้, ผล Test, ผล Recrawl และงานที่ยังเหลือ

กระบวนการทั้งหมดของคอร์สจึงจบที่สี่คำ: หลักฐาน การตัดสินใจ การแก้ไข และการยืนยันผลครับ
