# NoSQL Database - Vue.js + Vite Frontend

This project now includes a modern Vue.js frontend with Vite, while keeping the Go backend for the API.

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ 
- Go 1.21+
- npm or yarn

### Installation

1. **Install all dependencies:**
```bash
npm run install:all
```

2. **Start both frontend and backend:**
```bash
npm run dev
```

This will start:
- **Go Backend**: `http://localhost:8081` (API only)
- **Vue.js Frontend**: `http://localhost:5173` (Vite dev server)

## 📁 Project Structure

```
nosql-db/
├── frontend/                 # Vue.js + Vite application
│   ├── src/
│   │   ├── components/       # Vue components
│   │   ├── stores/          # Pinia stores
│   │   ├── views/           # Vue pages
│   │   └── router/          # Vue Router
│   ├── package.json
│   └── vite.config.ts       # Vite configuration
├── cmd/                     # Go backend
├── internal/                # Go database logic
├── config/                  # Configuration files
├── data/                    # Database files
└── package.json            # Root package.json
```

## 🎯 Available Scripts

### Root Level (Project Root)
```bash
npm run dev              # Start both frontend and backend
npm run dev:backend      # Start only Go backend
npm run dev:frontend     # Start only Vue.js frontend
npm run build            # Build Vue.js for production
npm run install:all      # Install all dependencies
```

### Frontend Only (frontend/ directory)
```bash
cd frontend
npm run dev              # Start Vite dev server
npm run build            # Build for production
npm run preview          # Preview production build
npm run lint             # Run ESLint
npm run format           # Format code with Prettier
```

## 🔧 Development

### Frontend Development
- **Vue 3** with Composition API
- **TypeScript** for type safety
- **Pinia** for state management
- **Vue Router** for navigation
- **Vite** for fast development and building

### Backend Development
- **Go** server with REST API
- **Transaction support** with WAL
- **JSON file storage**
- **Indexing system**

### API Communication
The Vue.js frontend communicates with the Go backend via:
- **Proxy**: Vite dev server proxies `/api/*` requests to `http://localhost:8081`
- **REST API**: All database operations go through the Go API
- **Transactions**: Full transaction support with begin/commit/rollback

## 🌐 Available Interfaces

1. **Vue.js Frontend**: `http://localhost:5173` (Modern, reactive UI)
2. **Classic HTML**: `http://localhost:8081` (Original template-based UI)
3. **Vue.js (served by Go)**: `http://localhost:8081/vue` (Alternative Vue.js UI)

## 🎨 Features

### Vue.js Frontend Features
- ✅ **Reactive UI** - Real-time updates
- ✅ **TypeScript** - Type safety
- ✅ **Modern Design** - Clean, responsive interface
- ✅ **Transaction Management** - Begin/Commit/Rollback
- ✅ **CRUD Operations** - Add/Edit/Delete documents
- ✅ **Modal Forms** - Clean form interfaces
- ✅ **Alert System** - Success/error notifications
- ✅ **Loading States** - Better user feedback

### Backend Features
- ✅ **REST API** - Full CRUD operations
- ✅ **Transaction Support** - ACID transactions with WAL
- ✅ **Indexing** - Unique and non-unique indexes
- ✅ **JSON Storage** - Document-based storage
- ✅ **Collections** - Multi-collection support

## 🚀 Production Deployment

### Build for Production
```bash
npm run build
```

This creates a `frontend/dist/` directory with optimized static files.

### Serve Production Build
You can serve the production build by:
1. Copying `frontend/dist/` contents to your web server
2. Configuring your web server to proxy `/api/*` to your Go backend
3. Or integrating the built files into your Go server

## 🔄 Development Workflow

1. **Start Development**: `npm run dev`
2. **Edit Frontend**: Modify files in `frontend/src/`
3. **Edit Backend**: Modify Go files in `cmd/` and `internal/`
4. **Hot Reload**: Both frontend and backend support hot reloading
5. **Build**: `npm run build` for production

## 📚 Technology Stack

### Frontend
- **Vue 3** - Progressive JavaScript framework
- **TypeScript** - Type-safe JavaScript
- **Vite** - Fast build tool and dev server
- **Pinia** - State management
- **Vue Router** - Client-side routing

### Backend
- **Go** - High-performance server language
- **JSON** - Document storage format
- **File System** - Simple, reliable storage
- **REST API** - Standard HTTP interface

## 🎯 Benefits of This Setup

1. **Separation of Concerns** - Frontend and backend are independent
2. **Modern Development** - Hot reload, TypeScript, modern tooling
3. **Scalability** - Can deploy frontend and backend separately
4. **Developer Experience** - Fast development with Vite
5. **Production Ready** - Optimized builds for production

## 🐛 Troubleshooting

### Common Issues

1. **Port Conflicts**: Make sure ports 8081 and 5173 are available
2. **CORS Issues**: Vite proxy handles this automatically in development
3. **Build Errors**: Check TypeScript errors in `frontend/src/`
4. **API Errors**: Check Go server logs for backend issues

### Debug Commands
```bash
# Check if backend is running
curl http://localhost:8081/api/collections

# Check if frontend is running
curl http://localhost:5173

# View backend logs
# (Check terminal where npm run dev:backend is running)

# View frontend logs
# (Check terminal where npm run dev:frontend is running)
``` 