# Lesson 7 — Read the Dashboard and Evidence

**Target duration:** 4 minutes

## Narration

[On screen: real summary, findings, inventory, and page-evidence screenshots.]

Begin with the audit summary, but never stop at the counts. Forty warnings may represent forty unrelated problems—or one broken template repeated forty times.

Filter the findings table by severity, rule, or URL. Error, warning, and information labels help prioritize review; they do not predict ranking loss. Open a finding to see its rule ID, version, evidence, remediation, and limitation.

Then inspect representative pages. The page view connects summary data to response status, title, description, canonical, robots directives, headings, images, hreflang, structured data, inlinks, outlinks, and associated issues.

[On screen: raw title beside rendered title.]

Raw and rendered evidence remain separate. Raw shows what the server returned before page JavaScript executed. Rendered shows the controlled browser result. If critical metadata appears only after JavaScript, that difference is meaningful evidence; rendered output should not silently replace the raw response.

Use this five-step investigation:

[On screen: Rule → representative URLs → evidence → root cause → limitation.]

First, identify the rule. Second, inspect representative affected URLs. Third, verify the observed evidence. Fourth, look for a shared component, template, or path. Fifth, read the limitation before recommending a change.

The objective is not to force every warning count to zero. Utility pages may intentionally be noindex. Decorative images may correctly use an empty alt value. A temporary failure may disappear on a retest. Professional quality comes from explaining which findings matter, why they matter, and how a fix will be verified.
