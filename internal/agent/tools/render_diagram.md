Renders a diagram as a temporary local HTTP page and returns the URL.

<usage>
- Provide diagram `format`, `content`, optional `title`, and optional `expire_after`
- Currently only `mermaid` format is supported
- The generated URL is bound to the current session and expires automatically
</usage>

<parameters>
- format: Diagram format to render. Only `mermaid` is supported.
- title: Optional page title shown above the rendered diagram.
- content: Diagram source text to render.
- expire_after: Optional TTL in seconds before the URL expires.
</parameters>

<notes>
- The tool requires a valid session context
- Rendered pages are served from a local ephemeral server
- Expired diagrams are no longer accessible
</notes>
