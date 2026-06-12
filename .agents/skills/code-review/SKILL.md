---
name: code-review
description: MUST be used before approving, merging, or publishing code changes. Reviews the current branch using git diffs and history, identifies bugs and regressions, checks project conventions, and provides an approval recommendation. Use this skill whenever the user asks to review code changes, review a branch, review before merge, or determine if code is ready for approval.
compatibility: opencode
---

# Code Review

I perform comprehensive reviews of the current branch and provide actionable feedback before merge.

## Review Focus Areas

* **Correctness** - Does the code work as intended?
* **Regressions** - Does this break existing functionality?
* **Maintainability** - Is the code readable and well-structured?
* **Project conventions** - Does this follow existing patterns?
* **Edge cases** - Are boundary conditions handled?
* **Missing tests** - Are there adequate tests for the changes?
* **Security concerns** - Are there vulnerabilities or unsafe patterns?
* **Performance concerns** - Are there inefficiencies or scalability issues?
* **Positive observations** - What good decisions should be preserved?

## When to Use This Skill

Use this skill when asked to:

* Review the current branch
* Review changes before merge
* Review a pull request
* Determine whether code is ready for approval
* Find bugs, regressions, and edge cases
* Assess overall change quality

**Important**: Always use this skill before approving, merging, or publishing code. Even if the user doesn't explicitly request a review, if they're asking about merging or approving changes, trigger this skill.

## Instructions

### Determine the Review Scope

1. **Identify the target branch**
   - Prefer `main` if it exists
   - Otherwise, identify the repository's default branch
   - If the user specifies a different target, use that

2. **Review only tracked changes**
   - Use `git diff` and `git log` to identify changes
   - Ignore ALL untracked files reported by `git status`
   - Focus on actual code changes, not generated files

3. **Include in your review**
   - Changes committed in the current branch that are not in the target branch
   - Staged changes
   - Unstaged modifications to tracked files

4. **Exclude from your review**
   - Generated files (unless intentionally modified)
   - Build artifacts
   - Vendor directories
   - Dependencies (unless intentionally modified)

### Review Process

For each reviewed file:

1. **Summarize what changed** - What was added, removed, or modified?
2. **Explain behavioral impact** - How does this change the program's behavior?
3. **Identify potential bugs or regressions** - What could go wrong?
4. **Highlight edge cases** - What boundary conditions might be missed?
5. **Point out convention violations** - What existing patterns are broken?
6. **Evaluate readability and maintainability** - Is this code clear?
7. **Consider security implications** - Are there vulnerabilities?
8. **Consider performance implications** - Are there inefficiencies?
9. **Identify missing tests** - What test coverage is needed?
10. **Suggest improvements** - What could be better?

Avoid speculative findings that are not reasonably supported by the code.

### Severity Levels

Use these severity levels for findings:

#### Critical
Likely to cause:
* Security vulnerabilities
* Data loss or corruption
* Production outages
* Severe functional defects

#### Major
Likely to cause:
* Incorrect behavior
* Significant regressions
* Reliability problems
* Important performance degradation
* Serious maintainability concerns

#### Minor
Issues involving:
* Small defects
* Inconsistencies
* Limited edge cases
* Readability concerns

#### Suggestion
Optional improvements such as:
* Refactorings
* Naming improvements
* Documentation updates
* Style recommendations

## Output Format

### Files Reviewed

List every reviewed file. For each file include:
* File path
* Change type: Committed, Staged, Unstaged, or Combination

### Findings

For each finding include:
* Severity (Critical, Major, Minor, Suggestion)
* File and location (line numbers if applicable)
* Explanation of the issue
* Recommended change

If no issues are identified, explicitly state: **No issues identified.**

### Positive Observations

Mention good changes worth preserving:
* Improved readability
* Better abstractions
* Simpler logic
* Useful safeguards
* Good test coverage
* Consistency with existing patterns

### Overall Assessment

Provide:
* Overall assessment of the branch
* Remaining risks
* Additional testing recommendations
* Merge recommendation using one of:
  * **Approve** - Ready to merge
  * **Approve with Minor Suggestions** - Can merge after addressing minor issues
  * **Request Changes** - Must address issues before merging

Explain the rationale behind your decision.

## Principles

* **Prioritize correctness over style** - Functionality matters more than formatting
* **Focus on actionable feedback** - Provide concrete, fixable recommendations
* **Respect existing project conventions** - Don't impose personal preferences
* **Do not invent problems without evidence** - Base findings on actual code issues
* **Recognize good engineering decisions** - Acknowledge well-done work
* **If the change is solid, say so** - Don't manufacture criticism for its own sake
