Asks the user a question and waits for their response. Use this tool when you need clarification, a decision between multiple options, or confirmation before proceeding with a potentially impactful action.

<when_to_use>
- You need to choose between multiple valid approaches with significant tradeoffs
- A requirement is genuinely ambiguous and you cannot infer the answer
- You need confirmation before a destructive or irreversible action
- The user's intent is unclear and guessing wrong would waste significant effort
</when_to_use>

<when_not_to_use>
- You can find the answer by searching the codebase
- The question is about code style (follow existing patterns instead)
- You're asking for permission to do what was already requested
- The answer can be reasonably inferred from context
- You're asking trivial yes/no questions you should decide yourself
</when_not_to_use>

<parameters>
- question (required): A clear, specific question. Provide enough context for the user to answer without re-reading the conversation.
- header (optional): Short label for the dialog title (max 30 chars). Keep it descriptive but concise.
- options (optional): A list of choices, each with a `label` (1-5 words) and optional `description`. If empty, a free text input is shown.
- multi (optional): Allow selecting multiple choices. Default: false.
- allow_text (optional): Allow typing a custom answer even when options are provided. Default: true when options exist.
</parameters>

<modes>
1. **Text input mode** (no options): Shows a text input field. User types freely and presses Enter.
2. **Options mode** (with options): Shows selectable options. User navigates with ↑/↓ and confirms with Enter. Number keys (1-9) also work for direct selection. When allow_text is true, user can also press 't' to switch to text input for a custom answer.
3. **Multi-select mode** (multi=true): User can toggle multiple options with Space, then confirm all with Enter.
</modes>

<tips>
- Be specific: "Should the retry logic use exponential backoff or fixed intervals?" not "How should I implement retries?"
- Provide options when possible — it's faster for the user to pick than to type
- Include relevant context in the question so the user doesn't need to scroll up
- After receiving the answer, proceed immediately without asking follow-up questions
- Don't use this tool more than 2-3 times per task
</tips>
