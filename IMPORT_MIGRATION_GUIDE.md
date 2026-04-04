# Import Migration Guide - MS-AI Backend

## Overview

This guide provides instructions for updating import paths to use the new Clean Architecture structure.

## Current State

The codebase now has **two parallel structures**:
- **Old Structure**: `internal/core/`, `internal/data/`, `internal/api/`, `pkg/`
- **New Structure**: `internal/domain/`, `internal/application/`, `internal/infrastructure/`, `internal/delivery/`, `internal/shared/`

Both structures currently work, but we need to migrate to the new structure exclusively.

## Import Path Mapping

### Domain Layer
```go
// OLD
import "github.com/835-droid/ms-ai-backend/internal/core/content/manga"
import "github.com/835-droid/ms-ai-backend/internal/core/user"

// NEW
import "github.com/835-droid/ms-ai-backend/internal/domain/manga"
import "github.com/835-droid/ms-ai-backend/internal/domain/user"
```

### Application Layer
```go
// OLD
import "github.com/835-droid/ms-ai-backend/internal/api/dto"

// NEW
import "github.com/835-droid/ms-ai-backend/internal/application/dtos"
import "github.com/835-droid/ms-ai-backend/internal/application/interfaces/repositories"
import "github.com/835-droid/ms-ai-backend/internal/application/interfaces/services"
```

### Infrastructure Layer
```go
// OLD
import "github.com/835-droid/ms-ai-backend/internal/data/content/manga"
import "github.com/835-droid/ms-ai-backend/internal/data/user"
import "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/mongo"
import "github.com/835-droid/ms-ai-backend/internal/data/infrastructure/postgres"

// NEW
import "github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/mongodb"
import "github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/postgres"
import "github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/hybrid"
import "github.com/835-droid/ms-ai-backend/internal/infrastructure/database"
import "github.com/835-droid/ms-ai-backend/internal/infrastructure/config"
```

### Delivery Layer
```go
// OLD
import "github.com/835-droid/ms-ai-backend/internal/api/handler"
import "github.com/835-droid/ms-ai-backend/internal/api/middleware"
import "github.com/835-droid/ms-ai-backend/internal/api/router"
import "github.com/835-droid/ms-ai-backend/pkg/response"

// NEW
import "github.com/835-droid/ms-ai-backend/internal/delivery/http/handlers"
import "github.com/835-droid/ms-ai-backend/internal/delivery/http/middleware"
import "github.com/835-droid/ms-ai-backend/internal/delivery/http/routers"
import "github.com/835-droid/ms-ai-backend/internal/delivery/http/responses"
```

### Shared Kernel
```go
// OLD
import "github.com/835-droid/ms-ai-backend/pkg/logger"
import "github.com/835-droid/ms-ai-backend/pkg/utils"
import "github.com/835-droid/ms-ai-backend/pkg/jwt"
import "github.com/835-droid/ms-ai-backend/pkg/i18n"
import "github.com/835-droid/ms-ai-backend/pkg/errors"
import "github.com/835-droid/ms-ai-backend/pkg/validator"

// NEW
import "github.com/835-droid/ms-ai-backend/internal/shared/logger"
import "github.com/835-droid/ms-ai-backend/internal/shared/utils"
import "github.com/835-droid/ms-ai-backend/internal/shared/jwt"
import "github.com/835-droid/ms-ai-backend/internal/shared/i18n"
import "github.com/835-droid/ms-ai-backend/internal/shared/errors"
import "github.com/835-droid/ms-ai-backend/internal/shared/utils/validator"
```

## Migration Strategy

### Step 1: Update Package Declarations
First, update the package declarations in the new files to match their new locations:

```bash
# Example for domain/manga files
# Change from: package manga
# To: package manga (stays the same, but in new location)
```

### Step 2: Update Import Paths in New Files
Update all import paths in the new structure files to reference other new structure files.

### Step 3: Update Import Paths in Old Files
Update all import paths in the old structure files to reference the new structure.

### Step 4: Test Incrementally
Test each layer after updating imports.

### Step 5: Remove Old Files
After all imports are updated and tested, remove the old structure.

## Automated Migration Script

Create a script `scripts/migrate_imports.sh`:

```bash
#!/bin/bash

# Import Migration Script for MS-AI Backend

echo "Starting import migration..."

# Find and replace import paths
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/core/content/manga|github.com/835-droid/ms-ai-backend/internal/domain/manga|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/core/user|github.com/835-droid/ms-ai-backend/internal/domain/user|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/api/dto|github.com/835-droid/ms-ai-backend/internal/application/dtos|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/data/content/manga|github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/mongodb|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/data/user|github.com/835-droid/ms-ai-backend/internal/infrastructure/persistence/mongodb|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/api/handler|github.com/835-droid/ms-ai-backend/internal/delivery/http/handlers|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/api/middleware|github.com/835-droid/ms-ai-backend/internal/delivery/http/middleware|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/internal/api/router|github.com/835-droid/ms-ai-backend/internal/delivery/http/routers|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/pkg/logger|github.com/835-droid/ms-ai-backend/internal/shared/logger|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/pkg/utils|github.com/835-droid/ms-ai-backend/internal/shared/utils|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/pkg/jwt|github.com/835-droid/ms-ai-backend/internal/shared/jwt|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/pkg/i18n|github.com/835-droid/ms-ai-backend/internal/shared/i18n|g' {} \;
find . -name "*.go" -type f -exec sed -i 's|github.com/835-droid/ms-ai-backend/pkg/errors|github.com/835-droid/ms-ai-backend/internal/shared/errors|g' {} \;

echo "Import migration completed!"
echo "Please review changes and test thoroughly."
```

## Manual Migration Checklist

### Domain Layer
- [ ] Update imports in `internal/domain/manga/*.go`
- [ ] Update imports in `internal/domain/user/*.go`

### Application Layer
- [ ] Update imports in `internal/application/dtos/*.go`
- [ ] Update imports in `internal/application/interfaces/repositories/*.go`
- [ ] Update imports in `internal/application/interfaces/services/*.go`

### Infrastructure Layer
- [ ] Update imports in `internal/infrastructure/persistence/mongodb/*.go`
- [ ] Update imports in `internal/infrastructure/persistence/postgres/*.go`
- [ ] Update imports in `internal/infrastructure/persistence/hybrid/*.go`
- [ ] Update imports in `internal/infrastructure/database/*.go`
- [ ] Update imports in `internal/infrastructure/config/*.go`

### Delivery Layer
- [ ] Update imports in `internal/delivery/http/handlers/**/*.go`
- [ ] Update imports in `internal/delivery/http/middleware/*.go`
- [ ] Update imports in `internal/delivery/http/routers/**/*.go`
- [ ] Update imports in `internal/delivery/http/responses/**/*.go`

### Shared Kernel
- [ ] Update imports in `internal/shared/logger/*.go`
- [ ] Update imports in `internal/shared/utils/*.go`
- [ ] Update imports in `internal/shared/jwt/*.go`
- [ ] Update imports in `internal/shared/i18n/*.go`
- [ ] Update imports in `internal/shared/errors/*.go`

### Entry Points
- [ ] Update imports in `cmd/server/main.go`
- [ ] Update imports in `cmd/create_admin/main.go`
- [ ] Update imports in `cmd/utils/gen_invite.go`

## Testing After Migration

### 1. Build Test
```bash
cd MS-AI
go build ./cmd/server
go build ./cmd/create_admin
go build ./cmd/utils
```

### 2. Unit Tests
```bash
go test ./internal/domain/...
go test ./internal/application/...
go test ./internal/infrastructure/...
go test ./internal/delivery/...
go test ./internal/shared/...
```

### 3. Integration Tests
```bash
go test ./test/...
```

### 4. Run Application
```bash
go run cmd/server/main.go
```

## Common Issues and Solutions

### Issue 1: Import Cycle
**Problem**: Circular dependencies between packages
**Solution**: Move shared types to a separate package or use interfaces

### Issue 2: Missing Imports
**Problem**: Import path doesn't exist
**Solution**: Verify the file exists in the new location and package name matches

### Issue 3: Package Name Mismatch
**Problem**: Package name doesn't match directory name
**Solution**: Update package declaration to match directory name

### Issue 4: Duplicate Definitions
**Problem**: Same type defined in multiple packages
**Solution**: Remove duplicates and use single source of truth

## Final Cleanup

After all imports are updated and tested:

```bash
# Remove old directories
rm -rf internal/core
rm -rf internal/data
rm -rf internal/api
rm -rf internal/container
rm -rf pkg

# Clean go mod
go mod tidy

# Final build test
go build ./...

# Final test
go test ./...
```

## Verification

After cleanup, verify:
1. ✅ All imports resolve correctly
2. ✅ Application builds successfully
3. ✅ All tests pass
4. ✅ Application runs correctly
5. ✅ No references to old paths remain

## Support

For issues during migration:
1. Check import paths match exactly
2. Verify package declarations
3. Ensure no circular dependencies
4. Review error messages carefully
5. Test incrementally

---

**Status**: Ready for import migration
**Next Step**: Update import paths systematically
**Goal**: Complete migration to new structure