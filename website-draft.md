# Art Portfolio Website Draft
## Domain: art.fahadfaruqi.com

---

## Design Vision

Swiss-inspired photo portfolio. Offset layouts, generous whitespace, strong typography. WebGL image grid with subtle shaders. Cinematic dark aesthetic, high contrast photos. Film-inspired grain + light leak effects. Horizontal infinite scroll for featured photos. Photos dominate, with technical metadata (camera, lens, settings) displayed in compact format.

---

## Core Features

### Image Display
- Support PNG, JPEG, WebP formats
- Images served through Cloudflare CDN at `https://assets.fahadfaruqi.com`
- API endpoint `https://assets.fahadfaruqi.com/api/metadata` provides image list with metadata
- Full-screen lightbox on click with keyboard/swipe navigation
- Responsive scaling (mobile to desktop)
- Preserve original photo aspect ratios (no fixed 3:2)
- WebGL image grid with subtle shader effects (light leak, film grain)
- Horizontal infinite scroll for featured photos

### Smart Preloading
- Native `loading="lazy"` for offscreen images
- Intersection Observer for smart preloading of near-viewport images
- Preload next 3-5 images before they enter viewport
- Low-res blur-up placeholders: generate small thumbnail on CDN upload or use CSS blur filter

### Image Grouping
- Set and number encoded in metadata (no filename convention required)
- Images can have any filename
- Groups extracted from metadata `set` field
- Ordered by metadata `number` field within each set
- Filterable by set in UI
- Group headers or tabs for navigation

### Theme Support
- System preference detection (prefers-color-scheme)
- Manual toggle override
- Smooth CSS transitions between themes
- Persisted in localStorage

### Transitions & Animations
- Framer Motion page transitions (crossfade between photo sets)
- Subtle parallax on scroll for images
- Gentle zoom on hover (not kinetic type like Stefan)
- Lightbox open/close transitions
- Staggered entry animations for grid items
- Infinite smooth scroll for featured photo section
- WebGL shader effects: displacement on photo switch, time-based background motion

---

## Tech Stack

### Framework
- **SvelteKit** (latest)
- **Svelte 5** (runes for reactivity)
- **TypeScript** for type safety

### Upload Script
- **Go** for upload script with EXIF extraction
- **AWS SDK v2 for Go** for R2/S3 operations
- **go-exif** library for EXIF data extraction

### Styling
- **Tailwind CSS** for utility classes
- Custom CSS for animations
- CSS Grid for responsive layouts
- CSS filters for global tone curve/color grading

### Components
- **svelte-bricks** (janosh/svelte-bricks v0.4+) for masonry layout (Svelte 5 compatible)
- **LiteLight** (@byronjohnson/litelight) for lightbox (touch/swipe, zero dependencies, ~4KB)
- **Framer Motion** for page transitions and animations
- **Three.js** + **@threlte/threlte** for WebGL image grid
- **Intersection Observer** action for preloading
- **Virtual scroll** for smooth infinite scroll

### Deployment
- Static export (adapter-static)
- GitHub Pages or Vercel/Cloudflare Pages
- Simple, fast, no build complexity
- Images hosted on Cloudflare R2, served via `https://assets.fahadfaruqi.com`

---

## File Structure

```
art.fahadfaruqi.com/
├── src/
│   ├── lib/
│   │   ├── components/
│   │   │   ├── GalleryGrid.svelte
│   │   │   ├── ImageCard.svelte
│   │   │   ├── Lightbox.svelte
│   │   │   ├── ThemeToggle.svelte
│   │   │   ├── FilterBar.svelte
│   │   │   ├── MetadataDisplay.svelte
│   │   │   └── ErrorPage.svelte
│   │   ├── utils/
│   │   │   ├── image-parser.ts
│   │   │   ├── preloader.ts
│   │   │   ├── metadata.ts
│   │   │   └── group-images.ts
│   │   └── stores.ts
│   ├── routes/
│   │   ├── +page.svelte
│   │   └── error/
│   │       └── +page.svelte
│   └── app.html
├── scripts/
│   ├── upload-images.go
│   └── go.mod
├── static/
├── .env.example
├── tailwind.config.js
├── svelte.config.js
└── package.json
```

---

## Component Architecture

### GalleryGrid.svelte
- Main container using svelte-bricks
- Filters images by active set
- Handles responsive breakpoints
- Manages image preloading queue
- Validates accessibility fields (title, altText, description) before rendering

### ImageCard.svelte
- Individual image wrapper
- Hover effects
- Click to open lightbox
- Lazy loading attributes
- Blur-up placeholder via CSS filter or small thumbnail from CDN
- CDN URL from metadata API response
- Render title and description under image
- Use MetadataDisplay component for EXIF data
- ARIA labels for accessibility

### Implementation
```svelte
<!-- ImageCard.svelte -->
<script lang="ts">
  import MetadataDisplay from './MetadataDisplay.svelte';
  
  export let image: ImageMetadata;
</script>

<div class="image-card">
  <img 
    src={image.url} 
    alt={image.altText}
    loading="lazy"
    class="image-card__image"
  />
  
  <div class="image-card__info">
    <h3 class="image-card__title">{image.title}</h3>
    <p class="image-card__description">{image.description}</p>
    <MetadataDisplay exif={image.exif} />
  </div>
</div>

<style>
  .image-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .image-card__image {
    width: 100%;
    height: auto;
    object-fit: cover;
    border-radius: 0.25rem;
  }
  
  .image-card__info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  .image-card__title {
    font-size: 1rem;
    font-weight: 600;
    margin: 0;
  }
  
  .image-card__description {
    font-size: 0.875rem;
    color: var(--text-secondary, #666);
    margin: 0;
  }
</style>
```

### Lightbox.svelte
- Full-screen overlay
- Previous/next navigation
- Keyboard support (esc, arrows)
- Smooth open/close animations
- Image info display + full EXIF data (camera, lens, settings, date)
- Use MetadataDisplay component (always expanded in lightbox)
- ARIA labels for screen readers
- Focus management (trap focus in lightbox)
- Live region announcements for image changes

### Implementation
```svelte
<!-- Lightbox.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import MetadataDisplay from './MetadataDisplay.svelte';
  
  export let images: ImageMetadata[];
  export let currentIndex = 0;
  export let isOpen = false;
  
  const dispatch = createEventDispatcher();
  
  const currentImage = $derived(images[currentIndex]);
  
  function handlePrevious() {
    if (currentIndex > 0) {
      currentIndex--;
    }
  }
  
  function handleNext() {
    if (currentIndex < images.length - 1) {
      currentIndex++;
    }
  }
  
  function handleClose() {
    dispatch('close');
  }
  
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') handleClose();
    if (e.key === 'ArrowLeft') handlePrevious();
    if (e.key === 'ArrowRight') handleNext();
  }
</script>

{#if isOpen}
  <div 
    class="lightbox" 
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-label="Image viewer"
  >
    <button class="lightbox__close" on:click={handleClose} aria-label="Close">
      ✕
    </button>
    
    <button 
      class="lightbox__nav lightbox__nav--prev" 
      on:click={handlePrevious}
      disabled={currentIndex === 0}
      aria-label="Previous image"
    >
      ←
    </button>
    
    <button 
      class="lightbox__nav lightbox__nav--next" 
      on:click={handleNext}
      disabled={currentIndex === images.length - 1}
      aria-label="Next image"
    >
      →
    </button>
    
    <div class="lightbox__content">
      <img 
        src={currentImage.url} 
        alt={currentImage.altText}
        class="lightbox__image"
      />
      
      <div class="lightbox__info">
        <h2 class="lightbox__title">{currentImage.title}</h2>
        <p class="lightbox__description">{currentImage.description}</p>
        <MetadataDisplay exif={currentImage.exif} expanded={true} />
      </div>
    </div>
  </div>
{/if}

<style>
  .lightbox {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.95);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  
  .lightbox__close {
    position: absolute;
    top: 1rem;
    right: 1rem;
    background: rgba(255, 255, 255, 0.1);
    border: none;
    color: white;
    font-size: 2rem;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 0.25rem;
  }
  
  .lightbox__nav {
    position: absolute;
    background: rgba(255, 255, 255, 0.1);
    border: none;
    color: white;
    font-size: 2rem;
    cursor: pointer;
    padding: 1rem;
    border-radius: 0.25rem;
  }
  
  .lightbox__nav--prev {
    left: 1rem;
  }
  
  .lightbox__nav--next {
    right: 1rem;
  }
  
  .lightbox__content {
    max-width: 90vw;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  
  .lightbox__image {
    max-width: 100%;
    max-height: 70vh;
    object-fit: contain;
  }
  
  .lightbox__info {
    margin-top: 1rem;
    color: white;
    text-align: center;
    max-width: 600px;
  }
  
  .lightbox__title {
    margin: 0 0 0.5rem 0;
  }
  
  .lightbox__description {
    margin: 0 0 1rem 0;
    opacity: 0.8;
  }
</style>
```

### ThemeToggle.svelte
- Sun/moon icon toggle
- System preference detection
- localStorage persistence
- Smooth theme transition

### FilterBar.svelte
- Extracted sets from metadata API response
- "All" option plus specific sets
- Active state styling
- Animation on filter change

### ErrorPage.svelte
- User-friendly error message
- Retry button for CDN connection
- Link to contact/support
- Graceful fallback UI

### MetadataDisplay.svelte
- Displays EXIF metadata in compact format
- Shows camera, lens, focal length, aperture, shutter, ISO, date
- Collapsible/expandable view to save space
- Clean typography for technical details
- Hidden if no EXIF data available
- Responsive layout (stacks on mobile)

### Implementation
```svelte
<!-- MetadataDisplay.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let exif: {
    camera?: string;
    lens?: string;
    focalLength?: string;
    aperture?: string;
    shutter?: string;
    iso?: number;
    date?: string;
  };
  
  let expanded = false;
  
  const hasExif = exif && Object.values(exif).some(v => v);
</script>

{#if hasExif}
  <div class="metadata-display">
    <button 
      class="metadata-toggle" 
      on:click={() => expanded = !expanded}
      aria-expanded={expanded}
    >
      <span class="metadata-icon">📷</span>
      <span class="metadata-label">
        {expanded ? 'Hide' : 'Show'} Camera Info
      </span>
    </button>
    
    {#if expanded}
      <div class="metadata-content">
        {#if exif.camera}
          <div class="metadata-item">
            <span class="metadata-key">Camera:</span>
            <span class="metadata-value">{exif.camera}</span>
          </div>
        {/if}
        {#if exif.lens}
          <div class="metadata-item">
            <span class="metadata-key">Lens:</span>
            <span class="metadata-value">{exif.lens}</span>
          </div>
        {/if}
        {#if exif.focalLength}
          <div class="metadata-item">
            <span class="metadata-key">Focal Length:</span>
            <span class="metadata-value">{exif.focalLength}</span>
          </div>
        {/if}
        {#if exif.aperture}
          <div class="metadata-item">
            <span class="metadata-key">Aperture:</span>
            <span class="metadata-value">{exif.aperture}</span>
          </div>
        {/if}
        {#if exif.shutter}
          <div class="metadata-item">
            <span class="metadata-key">Shutter:</span>
            <span class="metadata-value">{exif.shutter}</span>
          </div>
        {/if}
        {#if exif.iso}
          <div class="metadata-item">
            <span class="metadata-key">ISO:</span>
            <span class="metadata-value">{exif.iso}</span>
          </div>
        {/if}
        {#if exif.date}
          <div class="metadata-item">
            <span class="metadata-key">Date:</span>
            <span class="metadata-value">{exif.date}</span>
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .metadata-display {
    margin-top: 0.5rem;
    font-size: 0.75rem;
    color: var(--text-secondary, #666);
  }
  
  .metadata-toggle {
    background: none;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0;
    color: var(--text-secondary, #666);
    font-size: 0.75rem;
  }
  
  .metadata-toggle:hover {
    color: var(--text-primary, #333);
  }
  
  .metadata-content {
    margin-top: 0.5rem;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.25rem 0.5rem;
  }
  
  .metadata-key {
    font-weight: 500;
  }
  
  .metadata-value {
    font-family: monospace;
  }
</style>
```

---

## Metadata API Integration

### Fetching Image Metadata
```typescript
// metadata.ts
const METADATA_API = "https://assets.fahadfaruqi.com/api/metadata";

export interface ImageMetadata {
  url: string;
  key: string;
  size: number;
  uploaded: string;
  etag: string;
  set: string; // Required for grouping
  number: number; // Required for ordering
  title: string; // Required for accessibility
  altText: string; // Required for accessibility
  description: string; // Required for accessibility
  exif?: {
    camera?: string;
    lens?: string;
    focalLength?: string;
    aperture?: string;
    shutter?: string;
    iso?: number;
    date?: string;
  };
  aspectRatio?: string;
  width?: number;
  height?: number;
}

export interface MetadataResponse {
  count: number;
  objects: ImageMetadata[];
}

export async function fetchMetadata(): Promise<ImageMetadata[]> {
  try {
    const response = await fetch(METADATA_API);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data: MetadataResponse = await response.json();
    
    // Set and number come directly from metadata (no filename parsing)
    return data.objects.filter(img => 
      // Validate required fields
      img.title && img.altText && img.description && img.set && img.number
    );
  } catch (error) {
    console.error("Failed to fetch metadata:", error);
    // Fallback to cached data or empty array
    return [];
  }
}
```

### Image Grouping from Metadata
```typescript
// group-images.ts
import type { ImageMetadata } from './metadata';

export interface ImageGroup {
  name: string;
  images: ImageMetadata[];
  latestUpload: string;
}

export function groupImagesBySet(images: ImageMetadata[]): ImageGroup[] {
  const groups = new Map<string, ImageMetadata[]>();
  
  images.forEach(img => {
    const set = img.set || 'uncategorized';
    if (!groups.has(set)) {
      groups.set(set, []);
    }
    groups.get(set)!.push(img);
  });
  
  // Convert to array and sort by latest upload
  return Array.from(groups.entries()).map(([name, images]) => ({
    name,
    images: images.sort((a, b) => 
      new Date(b.uploaded).getTime() - new Date(a.uploaded).getTime()
    ),
    latestUpload: images.reduce((latest, img) => 
      new Date(img.uploaded) > new Date(latest) ? img.uploaded : latest,
      images[0].uploaded
    ),
  })).sort((a, b) => 
    new Date(b.latestUpload).getTime() - new Date(a.latestUpload).getTime()
  );
}
```

---

## Image Naming Convention

```
No filename convention required.
Set and number encoded in metadata.
Images can have any filename.

Supported formats: .png, .jpg, .jpeg, .webp
Aspect ratio: Preserve original photo ratio (no standardization)

CDN URL format: https://assets.fahadfaruqi.com/<filename>

Examples:
https://assets.fahadfaruqi.com/mountain-sunset.jpg
https://assets.fahadfaruqi.com/city-street.png
https://assets.fahadfaruqi.com/portrait-001.webp
```



---

## Preloading Strategy

### Three-tier approach:

1. **Viewport images**: Load immediately
2. **Near-viewport**: Preload when user scrolls near
3. **Offscreen**: Use native lazy loading

```typescript
// preloader.ts
const CDN_BASE = "https://assets.fahadfaruqi.com";

export function preloadImage(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve();
    img.onerror = reject;
    img.src = `${CDN_BASE}/${src}`;
  });
}

// Intersection Observer setup
const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement;
        preloadImage(img.dataset.src!);
        observer.unobserve(img);
      }
    });
  },
  { rootMargin: '200px' } // Preload 200px before viewport
);
```

---

## Styling Approach

### Color Palette
- **Dark mode default**: Deep black (#0a0a0a), light text (#f5f5f5)
- **Light mode optional**: Off-white (#fafafa), dark text (#1a1a1a)
- **Accent**: Subtle warm highlights for interactive elements
- **High contrast**: Photos pop against dark background

### Typography
- Clean sans-serif (Inter or system fonts)
- Large headings, readable body text
- Monospace font for technical metadata (camera settings)
- Minimal UI text - let images speak, but show technical details

### Layout
- Offset, asymmetric grid (Swiss print design influence)
- Photos bleed edges, create visual tension
- Masonry with responsive columns
- Mobile: 1-2 columns
- Tablet: 2-3 columns
- Desktop: 3-4 columns
- Gutter spacing: 16-24px
- Horizontal infinite scroll for featured photo section

---

## Image CDN Setup

### Cloudflare R2 + CDN
- Images stored in Cloudflare R2 bucket named `assets`
- Custom domain: `https://assets.fahadfaruqi.com`
- CDN caching enabled with proper headers
- Global edge distribution for fast loading

### Image Discovery API
- API endpoint: `https://assets.fahadfaruqi.com/api/metadata`
- Returns JSON with image metadata from R2 bucket:
  ```json
  {
    "count": 150,
    "objects": [
      {
        "url": "https://assets.fahadfaruqi.com/landscape-1.jpg",
        "key": "landscape-1.jpg",
        "size": 2048576,
        "uploaded": "2025-09-21T10:30:00Z",
        "etag": "\"abc123\"",
        "set": "landscape",
        "number": 1,
        "title": "Mountain Sunset",
        "altText": "Mountain landscape at sunset",
        "description": "A serene mountain landscape captured during golden hour",
        "exif": {
          "camera": "Sony A7IV",
          "lens": "24-70mm f/2.8",
          "focalLength": "35mm",
          "aperture": "f/2.8",
          "shutter": "1/250s",
          "iso": 100,
          "date": "2025-09-21"
        },
        "aspectRatio": "3:2",
        "width": 2400,
        "height": 1600
      }
    ]
  }
  ```
- API response cached at edge for 7 days, browser for 1 day
- Custom metadata fields spread into each object
- **Required accessibility fields**: title, altText, description (enforced on upload)
- Optional fields: set, number, exif, aspectRatio, width, height
- Parse `key` to extract set and number using `<set>-n.<ext>` convention
- Sort images by `uploaded` date (newest first)
- Fallback to cached data if API fails

### R2 Configuration
- Bucket name: `assets`
- Custom domain: `assets.fahadfaruqi.com`
- Worker route: `assets.fahadfaruqi.com/api/*`
- CORS enabled: `Access-Control-Allow-Origin: *`
- Edge cache: 7 days, browser cache: 1 day
- Custom metadata fields required on upload: title, altText, description, set, number
- Optional metadata fields: exif, aspectRatio, width, height

### DNS and SSL Setup
- DNS configured for `assets.fahadfaruqi.com`
- SSL certificate via Cloudflare SSL
- CNAME record pointing to R2/CDN
- Worker deployed to `metadata-api` on `fahadfaruqi.com` zone

### Environment Configuration
```bash
# No environment variables needed - CDN URL is fixed
# Metadata API: https://assets.fahadfaruqi.com/api/metadata
```

### Upload Workflow
1. Upload images to R2 `assets` bucket (any filename allowed)
2. **Required metadata fields**: title, altText, description, set, number (enforced on upload)
3. Optional metadata fields: exif, aspectRatio, width, height
4. Images auto-served via `https://assets.fahadfaruqi.com`
5. Metadata API auto-updates (edge cache invalidates on redeploy or after 7 days)
6. No image files in repo (keeps repo lightweight)
7. Version control only for code, not assets

### Upload Script/Interface
- Script or web interface enforces required accessibility fields before upload
- Prompts for title, altText, description for each image
- Validates fields are not empty before accepting upload
- Optionally prompts for optional metadata (set, number, exif data)
- Uploads to R2 with custom metadata attached
- Error handling for missing required fields

### Image Optimization Pipeline
- Automated PNG/JPEG/WebP optimization on upload
- Compression level 80-85
- Preserve original aspect ratios
- **Accessibility validation**: title, altText, description, set, number required before upload accepted
- Upload script enforces these fields as prerequisite

---

## Error Handling

### Metadata API Failures
- Try-catch around metadata fetch
- Fallback to cached metadata if API fails
- Retry mechanism with exponential backoff
- Graceful degradation: show empty state or cached images
- Global error boundary with retry button
- Validate required accessibility fields (title, altText, description) before rendering

### Image Loading Failures
- Health check on individual image URLs
- Graceful degradation: show error icon or gray placeholder for failed images
- Retry failed images on user interaction
- Log errors for monitoring

### Error Page
- Dedicated `/error` route for persistent failures
- Shows user-friendly message
- "Retry" button to reattempt metadata fetch
- Link to contact/support

---

## Performance Optimizations

1. **Image optimization**
   - PNG/JPEG/WebP formats
   - Compression level 80-85
   - Cloudflare CDN caching headers
   - Global edge distribution

2. **Code splitting**
   - Lazy load lightbox component
   - Dynamic imports for heavy libs

3. **Build optimization**
   - Static generation
   - CSS purging via Tailwind
   - Minification

4. **Runtime**
   - Debounced scroll handlers
   - RequestAnimationFrame for animations
   - Virtualization for large galleries

---

## WebGL Implementation

### Image Grid with Shaders
- Three.js + @threlte/threlte for WebGL integration
- Images rendered as textures on plane meshes
- Shared plane geometry between meshes for optimization
- Light leak/bokeh overlay shader (not LED like Stefan)
- Subtle film grain shader for cinematic effect
- Displacement effect on photo switch using noise-based pattern
- Time-based background motion pattern
- Mobile fallback: native HTML5 images (no WebGL for performance)

### Shader Effects Architecture
```typescript
// Vertex shader: displacement + curved bend
uniform float uTime;
uniform float uProgress;
varying vec2 vUv;

void main() {
  vUv = uv;
  vec3 pos = position;
  
  // Noise-based displacement on x-axis
  float noise = snoise(vec2(pos.x * 2.0, uTime));
  pos.x += noise * uProgress * 0.5;
  
  // Curved bend for depth
  pos.z += sin(pos.x * 0.5) * 0.1;
  
  gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
}
```

```typescript
// Fragment shader: light leak + film grain
uniform sampler2D uTexture;
uniform float uTime;
varying vec2 vUv;

void main() {
  vec4 color = texture2D(uTexture, vUv);
  
  // Light leak overlay
  float leak = smoothstep(0.3, 0.7, vUv.x + sin(uTime * 0.5) * 0.1);
  color.rgb += vec3(0.1, 0.05, 0.0) * leak;
  
  // Film grain
  float grain = random(vUv + uTime) * 0.05;
  color.rgb += grain;
  
  gl_FragColor = color;
}
```

---

## Virtual Scroll Implementation

### Infinite Smooth Scroll
- Custom scroll engine based on Virtual scroll principles
- Lethargy scroll for proper scroll event detection
- Content segmented into sections with height/top offset stored
- Only visible sections translated in frame loop (optimal DOM manipulation)
- Touch and resize support
- Works with content of any shape/size
- Used for featured photo horizontal scroll section

```typescript
// Virtual scroll controller
class VirtualScroll {
  sections: Section[] = [];
  scrollProgress = 0;
  
  update(progress: number) {
    this.scrollProgress = progress;
    const totalHeight = this.getTotalHeight();
    
    this.sections.forEach(section => {
      const sectionProgress = (progress * totalHeight - section.top) / section.height;
      if (this.isSectionVisible(sectionProgress)) {
        section.element.style.transform = `translateY(${sectionProgress * 100}%)`;
      }
    });
  }
  
  isSectionVisible(progress: number): boolean {
    return progress >= 0 && progress <= 1;
  }
}
```

---

## Upload Script Implementation

### Purpose
Enforce accessibility metadata requirements before uploading images to R2 bucket.

### Required Metadata Schema
```json
{
  "required": {
    "title": "string - Image title for display",
    "altText": "string - Alt text for screen readers",
    "description": "string - Detailed description of image content",
    "set": "string - Group/category name (e.g., 'landscape', 'portrait')",
    "number": "number - Sequential number within set for ordering"
  },
  "optional": {
    "exif": {
      "camera": "string - Camera model (e.g., 'Sony A7IV')",
      "lens": "string - Lens (e.g., '24-70mm f/2.8')",
      "focalLength": "string - Focal length (e.g., '35mm')",
      "aperture": "string - Aperture (e.g., 'f/2.8')",
      "shutter": "string - Shutter speed (e.g., '1/250s')",
      "iso": "number - ISO setting (e.g., 100)",
      "date": "string - Photo date (e.g., '2025-09-21')"
    },
    "aspectRatio": "string - Aspect ratio (e.g., '3:2', '16:9')",
    "width": "number - Image width in pixels",
    "height": "number - Image height in pixels"
  }
}
```

### Example Metadata
```json
{
  "title": "Mountain Sunset",
  "altText": "Mountain landscape at sunset with orange and purple sky",
  "description": "A serene mountain landscape captured during golden hour, showing layered peaks against a vibrant orange and purple sky with wispy clouds.",
  "set": "landscape",
  "number": 1,
  "exif": {
    "camera": "Sony A7IV",
    "lens": "24-70mm f/2.8",
    "focalLength": "35mm",
    "aperture": "f/2.8",
    "shutter": "1/250s",
    "iso": 100,
    "date": "2025-09-21"
  },
  "aspectRatio": "3:2",
  "width": 2400,
  "height": 1600
}
```

### Script Features
- Extracts EXIF data from image files automatically using `go-exif` library
- Auto-fills optional metadata fields from embedded EXIF
- Interactive CLI prompts for required fields (title, altText, description, set, number)
- Set and number required for grouping and ordering
- Batch upload support with smart defaults (uses previous set as default for subsequent files)
- Validates all required fields before upload with retry prompts
- Uploads to R2 using AWS SDK v2 for Go
- Error handling and retry logic

### Implementation
```go
// scripts/upload-images.go
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type UploadOptions struct {
	Title       string
	AltText     string
	Description string
	Set         string
	Number      int
	Exif        ExifData
	AspectRatio string
	Width       int
	Height      int
}

type ExifData struct {
	Camera      string
	Lens        string
	FocalLength string
	Aperture    string
	Shutter     string
	ISO         int
	Date        string
}

// Extract EXIF data from image file
func extractExifData(filePath string) (*UploadOptions, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	options := &UploadOptions{}

	// Parse EXIF data using simpler approach
	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		// Return empty options if EXIF not found (not critical)
		return options, nil
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return options, nil
	}

	ti := exif.NewTagIndex()
	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return options, nil
	}

	ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdPathStandard)
	if err != nil {
		return options, nil
	}

	// Extract camera info
	if make, err := ifd.TagByName("Make"); err == nil {
		if model, err := ifd.TagByName("Model"); err == nil {
			options.Exif.Camera = fmt.Sprintf("%s %s", make.Value(), model.Value())
		}
	}

	// Extract lens
	if lens, err := ifd.TagByName("LensModel"); err == nil {
		options.Exif.Lens = lens.Value()
	}

	// Extract focal length
	if focal, err := ifd.TagByName("FocalLength"); err == nil {
		options.Exif.FocalLength = fmt.Sprintf("%smm", focal.Value())
	}

	// Extract aperture
	if aperture, err := ifd.TagByName("FNumber"); err == nil {
		options.Exif.Aperture = fmt.Sprintf("f/%s", aperture.Value())
	}

	// Extract shutter speed
	if exposure, err := ifd.TagByName("ExposureTime"); err == nil {
		exposureTime := exposure.Value().(float64)
		if exposureTime > 0 {
			options.Exif.Shutter = fmt.Sprintf("1/%ds", int(1.0/exposureTime))
		}
	}

	// Extract ISO
	if iso, err := ifd.TagByName("ISOSpeedRatings"); err == nil {
		options.Exif.ISO = int(iso.Value().(uint16))
	}

	// Extract date
	if date, err := ifd.TagByName("DateTimeOriginal"); err == nil {
		options.Exif.Date = date.Value()
	}

	// Extract dimensions
	if width, err := ifd.TagByName("PixelXDimension"); err == nil {
		options.Width = int(width.Value().(uint32))
	}
	if height, err := ifd.TagByName("PixelYDimension"); err == nil {
		options.Height = int(height.Value().(uint32))
	}
	if options.Width > 0 && options.Height > 0 {
		options.AspectRatio = fmt.Sprintf("%d:%d", options.Width, options.Height)
	}

	return options, nil
}

func uploadImage(filePath string, options *UploadOptions, client *s3.Client, bucket string) error {
	// Validate required fields
	if options.Title == "" || options.AltText == "" || options.Description == "" {
		return fmt.Errorf("missing required accessibility fields: title, altText, description")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	key := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Determine content type
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}

	// Upload with metadata
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"title":       options.Title,
			"altText":     options.AltText,
			"description": options.Description,
			"set":         options.Set,
			"number":      strconv.Itoa(options.Number),
			"aspectRatio": options.AspectRatio,
			"width":       strconv.Itoa(options.Width),
			"height":      strconv.Itoa(options.Height),
			"camera":      options.Exif.Camera,
			"lens":        options.Exif.Lens,
			"focalLength": options.Exif.FocalLength,
			"aperture":    options.Exif.Aperture,
			"shutter":     options.Exif.Shutter,
			"iso":         strconv.Itoa(options.Exif.ISO),
			"date":        options.Exif.Date,
		},
	})

	return err
}

func promptForMetadata(filePath string, exifData *UploadOptions) (*UploadOptions, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nExtracted EXIF data for %s:\n", filepath.Base(filePath))
	if exifData.Exif.Camera != "" {
		fmt.Printf("  Camera: %s\n", exifData.Exif.Camera)
	}
	if exifData.Exif.Lens != "" {
		fmt.Printf("  Lens: %s\n", exifData.Exif.Lens)
	}
	if exifData.Exif.FocalLength != "" {
		fmt.Printf("  Focal Length: %s\n", exifData.Exif.FocalLength)
	}
	if exifData.Exif.Aperture != "" {
		fmt.Printf("  Aperture: %s\n", exifData.Exif.Aperture)
	}
	if exifData.Exif.Shutter != "" {
		fmt.Printf("  Shutter: %s\n", exifData.Exif.Shutter)
	}
	if exifData.Exif.ISO > 0 {
		fmt.Printf("  ISO: %d\n", exifData.Exif.ISO)
	}
	if exifData.Width > 0 && exifData.Height > 0 {
		fmt.Printf("  Dimensions: %dx%d\n", exifData.Width, exifData.Height)
	}

	fmt.Printf("Title for %s: ", filepath.Base(filePath))
	title, _ := reader.ReadString('\n')
	exifData.Title = strings.TrimSpace(title)

	fmt.Print("Alt text: ")
	altText, _ := reader.ReadString('\n')
	exifData.AltText = strings.TrimSpace(altText)

	fmt.Print("Description: ")
	description, _ := reader.ReadString('\n')
	exifData.Description = strings.TrimSpace(description)

	fmt.Print("Set (optional): ")
	set, _ := reader.ReadString('\n')
	exifData.Set = strings.TrimSpace(set)

	fmt.Print("Number (optional): ")
	numberStr, _ := reader.ReadString('\n')
	if numberStr != "" {
		if number, err := strconv.Atoi(strings.TrimSpace(numberStr)); err == nil {
			exifData.Number = number
		}
	}

	return exifData, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run upload-images.go <file1> [file2...]")
	}

	// Initialize R2 client using AWS SDK v2
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://<account-id>.r2.cloudflarestorage.com")
	})
	bucket := "assets"

	for _, filePath := range os.Args[1:] {
		exifData, err := extractExifData(filePath)
		if err != nil {
			log.Printf("Warning: failed to extract EXIF from %s: %v", filePath, err)
			exifData = &UploadOptions{}
		}

		options, err := promptForMetadata(filePath, exifData)
		if err != nil {
			log.Fatalf("Failed to prompt for metadata: %v", err)
		}

		if err := uploadImage(filePath, options, client, bucket); err != nil {
			log.Fatalf("Failed to upload %s: %v", filePath, err)
		}

		fmt.Printf("Uploaded %s successfully\n", filepath.Base(filePath))
	}
}
```

### Usage
```bash
# Initialize Go module
cd scripts
go mod init upload-images

# Install dependencies
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/dsoprea/go-exif/v3

# Set R2 credentials as environment variables
export R2_ACCESS_KEY_ID="your-access-key"
export R2_SECRET_ACCESS_KEY="your-secret-key"

# Update BaseEndpoint in script with your R2 account ID (replace <account-id>)

# Single image
go run upload-images.go ../images/landscape-1.jpg

# Batch upload
go run upload-images.go ../images/*.jpg
```

---

## User Flow

1. **Upload**: Upload images with required metadata (title, altText, description, set, number) enforced via Go script
2. **Landing**: Fetch metadata from API, render offset grid layout, all sets visible, WebGL image grid with subtle shaders
3. **Featured photos**: Horizontal infinite scroll section at top
4. **Filter**: Click set name to filter grid with Framer Motion crossfade
5. **View**: Click image to open lightbox with full EXIF data, title, and description
6. **Explore**: Toggle metadata display on image cards to see camera, lens, settings
7. **Navigate**: Use arrows or swipe in lightbox
8. **Theme**: Toggle dark/light mode (dark default for cinematic feel)
9. **Return**: Click outside image or press esc
10. **Scroll**: Smooth virtual scroll for featured section, WebGL + DOM hybrid

---

## Development Plan

### Phase 1: Foundation
- SvelteKit project setup
- Tailwind CSS configuration
- Basic gallery grid with offset layout
- Image loading system + CDN integration
- Metadata API integration (fetch from `https://assets.fahadfaruqi.com/api/metadata`)
- Image grouping logic from metadata
- EXIF data parsing and display
- Go upload script development (enforces title, altText, description)

### Phase 2: Core Features
- Filter bar component with Framer Motion crossfade
- Lightbox component with EXIF display
- Theme toggle (dark default)
- Three.js + @threlte/threlte WebGL setup

### Phase 3: WebGL & Advanced Interactions
- WebGL image grid with plane meshes
- Shader implementation (light leak, film grain, displacement)
- Virtual scroll engine for infinite horizontal scroll
- Page transitions with Framer Motion
- Mobile WebGL fallback (native HTML5)

### Phase 4: Polish & Performance
- Shader optimization (shared geometry, mesh skipping)
- Preloading optimization
- Mobile responsiveness
- Performance tuning (frame loop optimization)
- CSS tone curve/color grading

### Phase 5: Deploy
- Static build configuration
- Deployment to hosting (Vercel/Cloudflare Pages)
- Domain setup
- Testing across devices

---

## Success Metrics

- **Performance**: Lighthouse score 90+
- **UX**: Smooth scrolling, no visible image loading lag
- **Accessibility**: Keyboard navigation, screen reader support, ARIA labels, focus management
- **Responsiveness**: Works 320px - 4K
- **Maintainability**: Clean code, easy to add new images

---

## References & Inspiration

- **Stefan Vitasović Portfolio 2025**: Swiss print design, WebGL grid, virtual scroll, shader effects
- **Codrops case study**: Detailed breakdown of Stefan's technical implementation
- **Three.js + @threlte/threlte**: WebGL integration for Svelte
- **Framer Motion**: Page transitions and animations
- **Virtual scroll**: Smooth scrolling engine implementation
- **CSS-Tricks**: Lazy loading in Svelte patterns
- **svelte-bricks**: Masonry layout component
