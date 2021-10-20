From: <{MAIL_FROM}>
To: <{RCPT_TO}>
Subject: {SUBJECT}
Date: {TIME}
X-Mailer: {VERSION}
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="mailmaster-{BOUNDARY_LINE}"


--mailmaster-{BOUNDARY_LINE}
Content-Type: text/html; charset="utf-8"
Content-Transfer-Encoding: base64

{CONTENT}
--mailmaster-{BOUNDARY_LINE}--