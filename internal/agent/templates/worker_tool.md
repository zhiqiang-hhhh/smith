Launch a worker agent that can read, write, edit files, and execute commands to complete implementation tasks independently. Use this when you need to delegate a self-contained coding task that involves modifying files.

<when_to_use>
- Implementation tasks that can be done independently (e.g., "add error handling to this function", "write tests for this module")
- Refactoring a specific file or module
- Tasks that don't conflict with what you're currently doing (different files/modules)
- Multiple independent implementation tasks that can run in parallel
- When you want to offload work to keep your own context clean
</when_to_use>

<when_not_to_use>
- Simple searches or lookups — use the Agent tool instead (read-only, cheaper)
- Tasks that require back-and-forth discussion with the user
- Tasks that would modify the same files you're currently editing (conflict risk)
- Trivial single-line changes — do those yourself
</when_not_to_use>

<usage_notes>
1. Workers execute autonomously — they cannot ask you questions or see your conversation.
2. Write detailed, self-contained prompts. Include: what to change, which files, why, any constraints.
3. Workers have full tool access: bash, edit, write, grep, view, etc. They can run tests.
4. Workers auto-approve all permission prompts (no user interaction).
5. Launch multiple workers concurrently for independent tasks by using multiple tool calls in one message.
6. IMPORTANT: Avoid assigning overlapping file ranges to concurrent workers — they may conflict.
7. The worker's result is NOT visible to the user. Summarize the outcome for the user yourself.
8. Each worker invocation is stateless — you cannot send follow-up messages to a worker.
</usage_notes>

<writing_worker_prompts>
- Be specific about file paths and what changes to make
- Include relevant context the worker needs (architecture decisions, patterns to follow)
- Specify what verification to perform (run tests, check compilation)
- Example: "In /src/auth/login.go, add rate limiting to the Login handler. Use the existing RateLimiter from /src/middleware/rate.go. Add a test in login_test.go. Run `go test ./src/auth/`."
</writing_worker_prompts>
