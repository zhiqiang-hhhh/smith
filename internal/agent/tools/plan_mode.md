Switches between plan mode (read-only exploration and planning) and implementation mode.

<plan_mode>
When entering plan mode, you commit to ONLY using read-only tools (view, glob, grep, ls, agent, sourcegraph, web_search, fetch, diff) to explore the codebase, understand the problem, and formulate a detailed plan. You MUST NOT use any write tools (bash, edit, multiedit, write) while in plan mode.

Use plan mode for complex tasks where understanding the codebase first will prevent mistakes. A good plan includes:
- Which files need to change and why
- What the changes look like (pseudocode or description)
- What order to make changes in
- What tests to run to verify

After writing your plan, present it to the user and wait for approval before exiting plan mode.
</plan_mode>

<parameters>
- mode (required): Either "plan" to enter plan mode or "implement" to exit plan mode and begin implementation.
- plan (optional): When exiting plan mode, include the finalized plan that was approved by the user.
</parameters>

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
- Wait for user approval before switching to implement mode
- If the user rejects the plan, stay in plan mode and revise
</tips>
