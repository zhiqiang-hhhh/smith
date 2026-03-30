You are summarizing a conversation to preserve context for continuing work later.

**Critical**: This summary will be the ONLY context available when the conversation resumes. Assume all previous messages will be lost. Be thorough.

Before writing the summary, analyze the conversation chronologically in <analysis> tags:
1. Walk through each message/section and identify: user requests, your approach, key decisions, technical details (file names, code snippets, function signatures), errors encountered and how they were fixed
2. Pay special attention to specific user feedback — especially if they told you to do something differently
3. Double-check for technical accuracy and completeness

**Required sections**:

## Current State

### Task
- What task is being worked on (exact user request, quoted if possible)
- Broader context if the request is part of a larger effort

### Current Progress
- What's been completed so far
- What's being worked on right now (incomplete work)
- What remains to be done (specific next steps, not vague)

## Files & Changes

- Files that were modified (with brief description of changes and important line numbers)
- Files that were read/analyzed (why they're relevant)
- Key files not yet touched but will need changes
- Include full code snippets for critical changes that would be hard to reconstruct

## Technical Context

- Architecture decisions made and why
- Patterns being followed (with examples)
- Libraries/frameworks being used
- Commands that worked (exact commands with context)
- Commands that failed (what was tried and why it didn't work)
- Environment details (language versions, dependencies, etc.)

## Strategy & Approach

- Overall approach being taken
- Why this approach was chosen over alternatives
- Key insights or gotchas discovered
- Assumptions made
- Any blockers or risks identified

## All User Messages

List ALL non-tool-result user messages. These are critical for understanding the user's feedback and changing intent. Include them verbatim or closely paraphrased.

## Errors and Fixes

- Each error encountered, with:
  - What caused it
  - How it was fixed
  - Any user feedback on the fix

## Exact Next Steps

Be specific. Don't write "implement authentication" — write:

1. Add JWT middleware to src/middleware/auth.js:15
2. Update login handler in src/routes/user.js:45 to return token
3. Test with: npm test -- auth.test.js

IMPORTANT: Ensure the next step is DIRECTLY in line with the user's most recent explicit request. If the last task was concluded, only list next steps that are explicitly in line with the user's request. Do not start on tangential or old requests that were already completed. Include direct quotes from the most recent conversation showing exactly what task you were working on and where you left off.

**Tone**: Write as if briefing a teammate taking over mid-task. Include everything they'd need to continue without asking questions. No emojis ever.

**Length**: No limit. Err on the side of too much detail rather than too little. Critical context is worth the tokens.

## Key Facts (REQUIRED)

At the END of your summary, include a `<key_facts>` section with structured facts that will be auto-injected into every future prompt. Keep this under 500 tokens total. Use this exact format:

<key_facts>
- files_modified: list of files changed with one-line descriptions
- decisions: key technical decisions made and why
- user_preferences: any explicit preferences the user stated (code style, tools, approaches)
- errors_resolved: important errors that were fixed and how
- blocking_issues: any unresolved problems or blockers
- environment: relevant env details (versions, paths, configs)
- current_task: the exact task being worked on right now
</key_facts>

Each field should be a concise bullet list. Omit fields that have no relevant content. This section is machine-parsed — do not add commentary outside the tags.
