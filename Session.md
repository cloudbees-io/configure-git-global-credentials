# Code Review Session - PR #37: CBP-33337 Add V1 Deprecation Warning

## Date
March 23, 2026

## PR Details
- **Repository**: cloudbees-io/configure-git-global-credentials
- **Branch**: CBP_33337 → v1-stable
- **PR**: #37
- **Ticket**: CBP-33337 - Add V2 migration guide link to V1 action logs

## Review Comment Addressed

**Reviewer**: @cb-nsrinivasa (cloudbees-io/cbp-workflow-team)
**Comment Location**: `cmd/root.go:82-85`

```go
func doConfigure(command *cobra.Command, args []string) error {
    // Display v1 deprecation warning (only once)
    warningOnce.Do(printV1DeprecationWarning)
```

**Question**: "is sync.Once required for this? my understanding is doConfigure gets invoked once when configure cmd runs"

## Analysis Process

### Initial Analysis
1. Examined `cmd/root.go` structure:
   - `doConfigure` is a Cobra command handler (`RunE`)
   - Called once per program execution via `rootCmd.Execute()`
   - No goroutines calling `doConfigure` concurrently
   - Program exits after command completes

2. Checked execution flow:
   - `main()` → `cmd.Execute()` → `rootCmd.Execute()` → `doConfigure`
   - Single entry point, single execution path

### Deep Dive into CloudBees Platform
3. Examined actual deployment codebases in `/Users/sray/CBP_14218/CodeBases/`:
   - Reviewed workflow files using the action
   - Found examples in `platform-helmfiles`, `api`, `scm-service`, etc.

4. Analyzed action execution model from `action.yml`:
   ```yaml
   uses: docker://public.ecr.aws/l7o7z1g8/actions/configure-git-global-credentials:...
   args: "configure"
   entrypoint: "configure-git-global-credentials"
   ```

5. **Key Finding**: Each action step runs in a **separate Docker container**:
   - New container → New process → Runs once → Exits
   - No shared memory between workflow steps
   - Each `uses:` statement creates ephemeral container

### Conclusion
**Reviewer was correct**: `sync.Once` is unnecessary because:
- CLI tool runs once per Docker container invocation
- Process exits after single execution
- No concurrent goroutines
- No multiple invocations in same process
- Deprecation warning will print exactly once due to process lifecycle

`sync.Once` would only be needed if:
- Function called multiple times in single execution
- Concurrent goroutines might call it
- Long-running server/daemon process
- Binary used as a library with shared state

None of these apply to this CloudBees action.

## Changes Made

### Commit 1: `348b693 - Remove unnecessary sync.Once from doConfigure`

**File**: `cmd/root.go`

**Changes**:
1. Removed `"sync"` import (line 9)
2. Removed `warningOnce sync.Once` variable declaration (lines 29-30)
3. Changed function call (line 82):
   - FROM: `warningOnce.Do(printV1DeprecationWarning)`
   - TO: `printV1DeprecationWarning()`
4. Updated comment:
   - FROM: `// Display v1 deprecation warning (only once)`
   - TO: `// Display v1 deprecation warning`

**Git Diff**:
```diff
@@ -6,7 +6,6 @@ import (
 	"os"
 	"os/signal"
 	"strings"
-	"sync"
 	"syscall"

 	"github.com/cloudbees-io/configure-git-global-credentials/internal/configuration"
@@ -26,8 +25,6 @@ var (
 		Long:  "Configures the global git credentials",
 		RunE:  doConfigure,
 	}
-	// Ensure warning is printed only once
-	warningOnce sync.Once
 )

 func init() {
@@ -81,8 +78,8 @@ func cliContext() context.Context {
 }

 func doConfigure(command *cobra.Command, args []string) error {
-	// Display v1 deprecation warning (only once)
-	warningOnce.Do(printV1DeprecationWarning)
+	// Display v1 deprecation warning
+	printV1DeprecationWarning()

 	ctx := cliContext()
```

## Final Status
- ✅ Changes committed locally
- ✅ Changes pushed to remote branch `CBP_33337`
- ✅ PR #37 updated with new commit
- ⏳ Awaiting re-review from @cb-nsrinivasa

## Suggested GitHub Response

```markdown
Good catch! You're right - `sync.Once` is unnecessary here since `doConfigure` is only invoked once per process execution (each Docker container runs the binary once and exits). I've removed it in commit 348b693.
```

## Notes
- Same issue exists in `cloudbees-io/checkout` action (also has unnecessary `sync.Once`)
- Could be cleaned up there as well in future PR
- Pattern likely copied between actions without reconsidering necessity

## Related Files Reviewed
- `/Users/sray/CloudBees_Actions/configure-git-global-credentials/cmd/root.go`
- `/Users/sray/CloudBees_Actions/configure-git-global-credentials/cmd/helper.go`
- `/Users/sray/CloudBees_Actions/configure-git-global-credentials/action.yml`
- `/Users/sray/CloudBees_Actions/checkout/cmd/root.go`
- `/Users/sray/CBP_14218/CodeBases/platform-helmfiles/.cloudbees/workflows/workflow.yaml`
- `/Users/sray/CBP_14218/CodeBases/api/.cloudbees/workflows/workflow.yaml`
