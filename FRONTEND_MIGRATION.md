# Frontend Migration Guide

## Overview
The frontend has been completely redesigned using Vue 3 + TypeScript + Tailwind CSS 4.0 dashboard template.

## New Structure

### Technology Stack
- **Vue 3** - Composition API with `<script setup>`
- **TypeScript** - Full type safety
- **Tailwind CSS 4.0** - Modern utility-first CSS
- **Vite** - Fast build tool
- **Vue Router 4** - Client-side routing
- **Axios** - HTTP client for API calls
- **Lucide Icons** - Modern icon library

### Directory Structure
```
client/
├── src/
│   ├── api/
│   │   └── axios.ts           # Axios configuration
│   ├── assets/                # Images, fonts, etc
│   ├── components/
│   │   ├── common/            # Reusable components
│   │   ├── layout/            # Layout components (Sidebar, Header)
│   │   └── ...                # Other component categories
│   ├── icons/                 # Icon components
│   ├── router/
│   │   └── index.ts           # Route configuration
│   ├── views/
│   │   ├── Proxies.vue        # Main proxy management page
│   │   ├── Settings.vue       # Settings with notifications
│   │   └── FailureLogs.vue    # Failure history page
│   ├── App.vue                # Root component
│   └── main.ts                # Entry point
├── public/                    # Static assets
├── index.html                 # HTML template
├── package.json               # Dependencies
├── tailwind.config.js         # Tailwind configuration
├── tsconfig.json              # TypeScript configuration
└── vite.config.ts             # Vite configuration
```

## Installation

### Prerequisites
- Node.js 18+ (recommended 20+)
- npm, yarn, or pnpm

### Steps

1. **Install Node.js** (if not installed):
   ```bash
   # macOS (using Homebrew)
   brew install node

   # Or download from https://nodejs.org/
   ```

2. **Install dependencies**:
   ```bash
   cd client
   npm install
   # or
   yarn install
   # or
   pnpm install
   ```

3. **Add axios** (if not in package.json):
   ```bash
   npm install axios
   # or
   yarn add axios
   # or
   pnpm add axios
   ```

## Development

### Run development server:
```bash
npm run dev
# or
yarn dev
# or
pnpm dev
```

The app will be available at `http://localhost:5173` (Vite default port).

### Build for production:
```bash
npm run build
# or
yarn build
# or
pnpm build
```

This creates optimized files in `client/dist/` directory.

## New Pages

### 1. Proxies Page (`/`)
**Features:**
- View all proxies in a table
- Add new proxy manually
- Import proxies from file (format: `ip:port:username:password|name|contacts`)
- Edit proxy details
- Delete proxies
- Verify selected proxies (real-time SSE updates)
- Export selected proxies
- Search and filter
- Status indicators (Online/Offline)
- Real-time metrics (latency, speed, uptime)
- View failure logs for specific proxy

**Components:**
- Data table with sorting
- Modal forms for add/edit
- File upload for import
- Server-Sent Events for batch verification
- Bulk selection with checkboxes

### 2. Settings Page (`/settings`)
**Features:**
- General settings (URL, timeout, intervals)
- Authentication (username, password)
- SSL verification toggle
- Telegram notification configuration:
  - Enable/disable notifications
  - Bot token and chat ID
  - Notification types (down, recovery, IP change, IP stuck, low speed)
  - Low speed threshold
  - Daily summary settings
- Test notification button

**Components:**
- Form with validation
- Toggle switches for boolean settings
- Number inputs with min/max
- Time picker for daily summary
- Real-time settings save

### 3. Failure Logs Page (`/failure-logs`)
**Features:**
- View all failure events
- Filter by:
  - Proxy (dropdown)
  - Error type (ping_failed, speed_check_failed, ip_check_failed)
  - Date range
  - Page size
- Statistics panel (when proxy selected):
  - Total failures
  - Failures by type
  - Failure rate per day
- Pagination
- Click proxy name to filter
- Color-coded error types

**Components:**
- Filterable data table
- Statistics dashboard
- Pagination controls
- Date range picker

## API Integration

All pages use Axios to communicate with the Go backend:

### Endpoints Used:

**Proxies:**
- `GET /api/proxy` - List all proxies
- `POST /api/proxy` - Create proxy
- `PUT /api/proxy/:id` - Update proxy
- `DELETE /api/proxy/:id` - Delete proxy
- `GET /api/proxy/verify-batch?ids=...` - SSE for batch verification
- `POST /api/import` - Import proxies from file
- `GET /api/export/selected?ids=...` - Export selected proxies

**Settings:**
- `GET /api/settings` - Get current settings
- `PUT /api/settings` - Update settings
- `POST /api/testNotification` - Test Telegram notification

**Failure Logs:**
- `GET /api/failureLogs` - List failure logs with filters
- `GET /api/failureStats/:id` - Get statistics for specific proxy

### Authentication
The backend uses Basic Authentication. All requests will include credentials automatically.

## Features

### Dark Mode
- Automatically detects system preference
- Toggle in header
- Persisted in localStorage
- All components support dark mode

### Responsive Design
- Mobile-first approach
- Sidebar collapses on mobile
- Tables scroll horizontally on small screens
- Touch-friendly buttons and inputs

### Real-time Updates
- SSE for batch proxy verification
- Progress indicators during long operations
- Automatic polling (can be added if needed)

### User Experience
- Loading states for all async operations
- Error handling with user-friendly messages
- Confirmation dialogs for destructive actions
- Toast notifications (can be added)
- Keyboard shortcuts (future enhancement)

## Customization

### Colors
Edit `tailwind.config.js` or use CSS variables:
```css
:root {
  --color-primary: ...;
  --color-secondary: ...;
  --color-success: ...;
  --color-danger: ...;
  --color-warning: ...;
}
```

### Logo
Replace files in `public/images/logo/`:
- `logo.svg` - Light mode logo
- `logo-dark.svg` - Dark mode logo
- `logo-icon.svg` - Collapsed sidebar icon

### Branding
Update in `src/router/index.ts`:
```typescript
document.title = `${to.meta.title} | Your App Name`
```

## Build and Deploy

### Production Build:
```bash
cd client
npm run build
```

This creates `client/dist/` with:
- index.html
- assets/ (CSS, JS, images)

### Deploy with Go Backend:
The Go backend already serves `client/dist`:
```go
router.Use(static.Serve("/", static.LocalFile("./client/dist", true)))
```

Just rebuild and restart the Go server:
```bash
# Build frontend
cd client && npm run build && cd ..

# Build and run Go backend
cd code
go build -o ../proxycheck
cd ..
./proxycheck
```

Access at `http://localhost:8080`

## Troubleshooting

### Build Errors

**Issue: TypeScript errors**
```bash
npm run type-check
```
Fix type errors or set `"strict": false` in `tsconfig.json`

**Issue: Missing dependencies**
```bash
rm -rf node_modules package-lock.json
npm install
```

**Issue: Tailwind not working**
Check `postcss.config.js` and `tailwind.config.js` exist

### Runtime Errors

**Issue: API calls fail with CORS**
- Backend should not have CORS issues (same origin)
- Check backend is running on port 8080
- Check Basic Auth credentials

**Issue: Routes show 404**
- Backend should have catch-all route:
```go
router.NoRoute(func(c *gin.Context) {
  if !strings.HasPrefix(c.Request.RequestURI, "/api") {
    c.File("./client/dist/index.html")
  }
})
```

**Issue: Icons not showing**
- Check lucide-vue-next is installed: `npm install lucide-vue-next`
- Verify imports in components

**Issue: Styles not loading**
- Check dist/assets/ has CSS files
- Clear browser cache
- Rebuild: `npm run build`

## Migration Checklist

- [x] Copy Tailwind dashboard template
- [x] Create Proxies page with full CRUD
- [x] Create Settings page with notifications
- [x] Create Failure Logs page
- [x] Update router with new routes
- [x] Update sidebar menu
- [x] Configure axios
- [ ] Install Node.js (if needed)
- [ ] Install dependencies (`npm install`)
- [ ] Install axios (`npm install axios`)
- [ ] Run development server (`npm run dev`)
- [ ] Test all pages and features
- [ ] Build for production (`npm run build`)
- [ ] Test with Go backend
- [ ] Deploy

## Next Steps

1. **Install Dependencies:**
   ```bash
   cd /Users/artiom/.claude-worktrees/proxycheck/festive-noether/client
   npm install
   npm install axios
   ```

2. **Development:**
   ```bash
   npm run dev
   ```
   Open `http://localhost:5173` to test

3. **Production Build:**
   ```bash
   npm run build
   ```

4. **Start Backend:**
   ```bash
   cd /Users/artiom/.claude-worktrees/proxycheck/festive-noether
   ./proxychecker.exe
   ```
   Open `http://localhost:8080`

## Additional Features (Future)

### Recommended Enhancements:
1. **Toast Notifications** - Use vue-toastification
2. **Charts** - Add ApexCharts for speed/uptime graphs
3. **WebSocket** - Real-time proxy status updates
4. **CSV Export** - Export failure logs to CSV
5. **Bulk Operations** - Delete multiple proxies
6. **Proxy Groups** - Organize proxies by tags/groups
7. **Advanced Filters** - More filtering options
8. **User Management** - Multiple users with roles
9. **API Rate Limiting** - Frontend throttling
10. **Keyboard Shortcuts** - Power user features

## Support

For issues or questions:
1. Check this documentation
2. Review browser console for errors
3. Check Go backend logs
4. Verify API endpoints are working (use Postman)
5. Check network tab in browser DevTools

## Summary

The new frontend provides:
- ✅ Modern, professional UI/UX
- ✅ Full proxy management (CRUD operations)
- ✅ Telegram notification configuration
- ✅ Failure history and statistics
- ✅ Dark mode support
- ✅ Responsive design
- ✅ TypeScript type safety
- ✅ Fast Vite build system
- ✅ Clean, maintainable code structure
- ✅ Integration with all backend APIs

All backend features are now accessible through a beautiful, user-friendly interface!
