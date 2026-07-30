package application

import (
	"strings"
	"testing"
)

func TestRedactRenderedDOMRemovesFormAndURLSecrets(t *testing.T){t.Parallel();input:=[]byte(`<html><body><form action="/submit?token=secret"><input name="password" value="hunter2"><textarea>private note</textarea><script>const apiKey='secret'</script><a href="https://example.com/path?session=abc#part">link</a></form></body></html>`);output,err:=redactRenderedDOM(input);if err!=nil{t.Fatal(err)};text:=string(output);for _,secret:=range []string{"hunter2","private note","apiKey","token=secret","session=abc"}{if strings.Contains(text,secret){t.Fatalf("retained %q in %s",secret,text)}};if !strings.Contains(text,"[redacted]")||!strings.Contains(text,"[script omitted]"){t.Fatalf("redaction markers missing: %s",text)}}
