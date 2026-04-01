Switches between plan mode (read-only exploration and planning) and implementation mode.

<plan_mode>
When entering plan mode, you commit to ONLY using read-only tools (view, glob, grep, ls, agent, sourcegraph, web_search, fetch, diff) to explore the codebase, understand the problem, and formulate a detailed plan. You MUST NOT use any write tools (bash, edit, multiedit, write, download) while in plan mode.

Use plan mode for complex tasks where understanding the codebase first will prevent mistakes. A good plan includes:
- Which files need to change and why
- What the changes look like (pseudocode or description)
- What order to make changes in
- What tests to run to verify
</plan_mode>

<parameters>
- mode (required): Either "plan" to enter plan mode or "implement" to exit plan mode and begin implementation.
- plan (required when mode is "implement"): The complete, finalized implementation plan. The user will review this plan before approving. Include all details: files to change, what to change, and verification steps.
</parameters>

<workflow>
1. Enter plan mode by calling with mode="plan"
2. Use read-only tools extensively to explore the codebase
3. Use ask_user to clarify requirements or approach if needed
4. When your plan is ready, call with mode="implement" and include the COMPLETE plan in the plan parameter
5. The user will see your plan and choose to:
   - Approve: you proceed with implementation
   - Approve and clear context: conversation history is cleared, you start fresh with just the plan
   - Reject (with optional feedback): you stay in plan mode and revise based on their feedback
</workflow>

<when_to_use>
- Complex multi-file refactoring where understanding the full scope is critical
- When the user asks you to "plan first" or "think before coding"
- Architecture changes that affect many components
- When you're unsure about the right approach and want to explore first
</when_to_use>

<when_not_to_use>
- Simple single-file changes
- Bug fixes with obvious solutions
- Tasks where the implementation path is clear
- Follow-up changes to code you just wrote
</when_not_to_use>

<tips>
- Enter plan mode BEFORE making any changes
- Use read-only tools extensively while planning
- Present the plan clearly with file paths and descriptions
- Always include the full plan in the plan parameter when calling mode="implement"
- If the user rejects the plan, incorporate their feedback and try again
</tips>
