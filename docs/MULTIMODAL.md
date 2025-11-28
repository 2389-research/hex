# Multimodal (Vision) Support in Clem

Clem supports Claude's vision capabilities, allowing you to analyze images alongside text prompts. This feature enables screenshot analysis, diagram explanation, OCR, and more.

## Supported Formats

- PNG (`.png`)
- JPEG (`.jpg`, `.jpeg`)
- GIF (`.gif`)
- WebP (`.webp`)

## Size Limits

- Maximum file size: 5MB per image
- This limit is enforced by the Claude API

## Using Vision in CLI Mode

### Basic Image Analysis

Analyze a single image:

```bash
clem --print --image screenshot.png "What's in this image?"
```

### Multiple Images

Compare multiple images in one request:

```bash
clem --print --image diagram1.png --image diagram2.png "Compare these two diagrams"
```

### Image-Only Requests

You can send images without a text prompt:

```bash
clem --print --image photo.jpg
```

Claude will provide a general description of the image.

## Usage Examples

### Screenshot Analysis

```bash
clem --print --image error-screen.png "What's the error and how do I fix it?"
```

### Diagram Explanation

```bash
clem --print --image architecture.png "Explain this system architecture"
```

### OCR (Text Extraction)

```bash
clem --print --image document.jpg "Extract all text from this image"
```

### Code from Screenshot

```bash
clem --print --image code-screenshot.png "Convert this code to markdown"
```

### Data from Charts

```bash
clem --print --image chart.png "What are the key insights from this chart?"
```

## Interactive Mode (Future)

The `:attach` command will be available in interactive mode to attach images during a conversation. This feature is planned for a future release.

## API Considerations

### Token Costs

Vision requests use more tokens than text-only requests:

- Images consume tokens based on their size and resolution
- Typical image: ~1000-2000 tokens
- High-resolution images may use significantly more

### Pricing

Vision features are more expensive than text-only interactions. Check Anthropic's pricing page for current rates:
https://www.anthropic.com/pricing

### Best Practices

1. **Resize images** before sending if possible to reduce token usage
2. **Use appropriate quality**: Don't send 4K screenshots when 1080p will do
3. **Be specific** in your prompts to get better results
4. **Batch related images** together when comparing or analyzing multiple items

## Technical Details

### How It Works

When you use the `--image` flag:

1. Clem loads the image file and validates format/size
2. The image is base64-encoded
3. A multimodal message is constructed with both image and text content blocks
4. The request is sent to Claude's Messages API with vision support

### ContentBlock Format

Internally, multimodal messages use an array of content blocks:

```go
msg := core.Message{
    Role: "user",
    ContentBlock: []core.ContentBlock{
        core.NewImageBlock(&core.ImageSource{
            Type:      "base64",
            MediaType: "image/png",
            Data:      "...",
        }),
        core.NewTextBlock("What's in this image?"),
    },
}
```

### Backward Compatibility

Text-only messages continue to work exactly as before:

```go
msg := core.Message{
    Role:    "user",
    Content: "Hello, world!",
}
```

The API client automatically detects whether to use string content or content blocks based on the message structure.

## Error Handling

### Common Errors

**File not found:**
```
Error: load image screenshot.png: image file not found
```
Solution: Check the file path and ensure the file exists.

**Unsupported format:**
```
Error: load image document.pdf: unsupported image format: .pdf
```
Solution: Convert to PNG, JPEG, GIF, or WebP.

**File too large:**
```
Error: load image huge.png: image file too large: 6291456 bytes (max 5242880 bytes)
```
Solution: Resize or compress the image to under 5MB.

## Examples with Output

### Example 1: Screenshot Analysis

```bash
$ clem --print --image error-screenshot.png "What's the error?"

The screenshot shows a Python traceback with a KeyError on line 42 of main.py.
The error occurs because the code is trying to access a dictionary key 'config'
that doesn't exist. To fix this, you should either:

1. Check if the key exists before accessing it:
   if 'config' in data:
       config = data['config']

2. Use .get() with a default value:
   config = data.get('config', {})
```

### Example 2: Diagram Explanation

```bash
$ clem --print --image architecture.png "Explain this architecture"

This diagram shows a microservices architecture with:
- A load balancer distributing traffic to multiple API servers
- Three services: User Service, Order Service, and Payment Service
- Each service has its own database
- An event bus for inter-service communication
- A cache layer (Redis) for frequently accessed data

The architecture follows common microservices patterns with service isolation
and event-driven communication.
```

## Limitations

1. **No video support**: Only static images are supported
2. **No audio**: Audio/video files cannot be processed
3. **Quality matters**: Blurry or low-quality images may not analyze well
4. **Context window**: Large images count toward your context window limit
5. **Rate limits**: Subject to Anthropic API rate limits

## Tips for Best Results

1. **Clear images**: Use high-quality, well-lit images
2. **Crop appropriately**: Remove unnecessary parts of the image
3. **Multiple angles**: For complex objects, provide multiple views
4. **Combine with text**: Add specific questions or context in your prompt
5. **Iterate**: If results aren't good, try rephrasing your question

## Future Enhancements

Planned features for multimodal support:

- Interactive mode `:attach` command
- Image history in conversation context
- Automatic image compression/resizing
- Screenshot capture integration
- Clipboard image support
