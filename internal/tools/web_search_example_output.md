# WebSearch Tool Example Output

This document shows example output from the WebSearch tool.

## Example 1: Basic Search

**Parameters:**
```json
{
  "query": "golang testing"
}
```

**Output:**
```markdown
# Search Results for: golang testing

Found 5 results:

### 1. Example Page 1
**URL**: https://example.com/page1

This is the first example result snippet

---

### 2. Test Page 2
**URL**: https://test.org/page2

This is the second test result snippet

---

### 3. Example Page 3
**URL**: https://example.com/page3

This is the third example result snippet

---

### 4. Blocked Page 4
**URL**: https://blocked.com/page4

This should be blocked

---

### 5. Another Page 5
**URL**: https://another.net/page5

This is the fifth result snippet

---
```

## Example 2: Limited Results

**Parameters:**
```json
{
  "query": "golang testing",
  "limit": 2
}
```

**Output:**
```markdown
# Search Results for: golang testing

Found 2 results:

### 1. Example Page 1
**URL**: https://example.com/page1

This is the first example result snippet

---

### 2. Test Page 2
**URL**: https://test.org/page2

This is the second test result snippet

---
```

## Example 3: Domain Filtering (Allowed)

**Parameters:**
```json
{
  "query": "golang testing",
  "allowed_domains": ["example.com"]
}
```

**Output:**
```markdown
# Search Results for: golang testing

Found 2 results:

### 1. Example Page 1
**URL**: https://example.com/page1

This is the first example result snippet

---

### 2. Example Page 3
**URL**: https://example.com/page3

This is the third example result snippet

---
```

## Example 4: Domain Filtering (Blocked)

**Parameters:**
```json
{
  "query": "golang testing",
  "blocked_domains": ["blocked.com"]
}
```

**Output:**
```markdown
# Search Results for: golang testing

Found 4 results:

### 1. Example Page 1
**URL**: https://example.com/page1

This is the first example result snippet

---

### 2. Test Page 2
**URL**: https://test.org/page2

This is the second test result snippet

---

### 3. Example Page 3
**URL**: https://example.com/page3

This is the third example result snippet

---

### 4. Another Page 5
**URL**: https://another.net/page5

This is the fifth result snippet

---
```

## Example 5: No Results Found

**Parameters:**
```json
{
  "query": "nonexistent query that returns nothing"
}
```

**Output:**
```markdown
# Search Results for: nonexistent query that returns nothing

No results found.
```
