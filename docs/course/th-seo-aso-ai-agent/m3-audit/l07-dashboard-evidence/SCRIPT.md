# บทที่ 7 — อ่าน Dashboard จากตัวเลขไปถึงหลักฐาน

**เวลาเป้าหมาย:** 4 นาที

## บทพูด

[ภาพบนจอ: Summary, Findings, Page Inventory และ Evidence Drawer]

เวลาเปิดผล Audit เรามักถูกดึงสายตาไปที่ตัวเลขสีแดงก่อน แต่จำนวน Error ไม่ได้บอกว่างานไหนกระทบธุรกิจมากที่สุด Warning 200 รายการอาจเกิดจาก Component ตัวเดียว หรืออาจเป็น 200 ปัญหาคนละแบบก็ได้

เริ่มจาก Summary เพื่อเห็นภาพรวม จากนั้น Filter ตาม Severity, Rule หรือข้อความใน Evidence แล้วเปิด Finding ที่เป็นตัวแทนขึ้นมาดู Toad จะบอก Rule ID, Version, หลักฐานที่พบ วิธีแก้ และข้อจำกัดของกฎ

[ภาพบนจอ: ซูมที่ Rule ID, Evidence, Remediation และ Limitations]

ต่อไปเปิด Page Inventory เพื่อเชื่อม Finding กลับไปยัง URL จริง ดู Status, Title, Canonical, Robots, Heading, Inlink, Outlink, Hreflang, Image และ Structured Data

Raw กับ Rendered ต้องแยกกัน Raw คือ HTML ที่ Server ส่งมาตั้งแต่แรก ส่วน Rendered คือ DOM หลัง Browser รัน JavaScript แล้ว ถ้า Title โผล่มาเฉพาะหลัง JavaScript นั่นเป็นข้อมูลสำคัญ เราไม่ควรเอาผล Rendered ไปทับ Raw แล้วทำเหมือนไม่เคยมีความต่าง

ผมแนะนำขั้นตอนสั้น ๆ ห้าข้อ หนึ่ง ดู Rule สอง เลือก URL ตัวแทน สาม ตรวจ Evidence สี่ หา Root Cause ที่อาจอยู่ใน Template และห้า อ่าน Limitation ก่อนเสนอให้แก้

[ภาพบนจอ: Rule → URL ตัวแทน → Evidence → Root Cause → Limitation]

เป้าหมายไม่ใช่ทำให้ Warning เป็นศูนย์ หน้า Utility บางหน้าอาจตั้ง noindex ถูกต้อง รูปตกแต่งอาจใช้ alt ว่างอย่างถูกต้อง และ Error บางอย่างอาจเกิดชั่วคราว คนทำ Audit ที่ดีต้องอธิบายได้ว่าอะไรควรแก้ อะไรควรตรวจเพิ่ม และอะไรไม่ต้องทำครับ
