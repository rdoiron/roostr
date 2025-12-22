# Roostr: Handoff to Claude Code

This document explains how to transition from planning to building with Claude Code CLI.

---

## Step 1: Create the GitHub Repository

### Option A: GitHub CLI (Recommended)

```bash
# Navigate to the bootstrap folder
cd roostr-bootstrap

# Create the repo on GitHub and push
gh repo create roostr --public --source=. --push
```

### Option B: Manual

1. Go to github.com and create a new repo named `roostr`
2. Don't initialize with README (we have one)
3. Then push the bootstrap folder:

```bash
cd roostr-bootstrap
git init
git add .
git commit -m "Initial project structure"
git branch -M main
git remote add origin git@github.com:YOURUSERNAME/roostr.git
git push -u origin main
```

---

## Step 2: Clone to Your Working Directory

```bash
# Clone to wherever you do development
cd ~/projects  # or wherever
git clone git@github.com:YOURUSERNAME/roostr.git
cd roostr
```

---

## Step 3: Start Your First Claude Code Session

```bash
# In the roostr directory
claude
```

---

## Step 4: First Session Kickoff Prompt

Copy and paste this as your first message to Claude Code:

---

```
I'm starting development on Roostr, a private Nostr relay management app.

Please read these files to understand the project:
1. CLAUDE.md - Project overview and conventions
2. docs/TASKS.md - Development task checklist
3. docs/USER-GUIDE.md - Feature documentation
4. docs/API.md - API reference

After reading, let's start with the first task: SETUP-001 (Initialize git repo with .gitignore).

Actually, that's already done. Let's move to SETUP-002: Create Makefile with all targets.

Before we start coding, confirm you understand:
- The tech stack (Go backend, Svelte frontend)
- The folder structure
- The coding conventions

Then let's begin.
```

---

## Step 5: Session Workflow

Each Claude Code session should follow this pattern:

### Starting a Session

1. Claude Code auto-reads CLAUDE.md
2. Ask it to check docs/TASKS.md for current progress
3. Pick the next unchecked task
4. Work on it

### During a Session

- Focus on 1-3 related tasks
- Test as you go
- Commit working code frequently
- Update TASKS.md checkboxes when done

### Ending a Session

```
Let's wrap up this session. Please:
1. Make sure all changes are saved
2. Update docs/TASKS.md with completed tasks
3. Commit with a descriptive message
4. Tell me what we accomplished and what's next
```

---

## Step 6: Useful Claude Code Commands

Inside Claude Code, you can run these:

```bash
# Check project structure
ls -la

# Run the API (once set up)
make api

# Run the UI (once set up)
make ui

# Run tests
make test

# Check what tasks are done
grep "\[x\]" docs/TASKS.md | wc -l
```

---

## Key Files Claude Code Should Know About

| File | Purpose | When to Read |
|------|---------|--------------|
| CLAUDE.md | Project overview, conventions | Auto-read on session start |
| docs/TASKS.md | Task checklist | Start of each session |
| docs/USER-GUIDE.md | Feature documentation | When implementing features |
| docs/API.md | API reference | When working on endpoints |
| app/api/go.mod | Go dependencies | When adding packages |
| app/ui/package.json | UI dependencies | When adding packages |

---

## Tips for Effective Sessions

### Do:
- Start with small, focused tasks
- Test each piece before moving on
- Commit often with clear messages
- Reference the spec when implementing features
- Ask Claude to explain before coding if unclear

### Don't:
- Try to build too much in one session
- Skip testing
- Forget to update TASKS.md
- Let context get stale (commit and restart if confused)

---

## Sample Prompts for Common Tasks

### Starting a new feature
```
Let's implement the [feature name] feature.
Please read docs/USER-GUIDE.md and docs/API.md first,
then let's plan the implementation before coding.
```

### Debugging
```
The [component] isn't working. The error is: [error message]
Please check [file] and help me fix it.
```

### Reviewing progress
```
Let's review what we've built so far. 
Check docs/TASKS.md and summarize our progress.
What should we focus on next?
```

### Ending the day
```
Let's commit our progress. Please:
1. Update docs/TASKS.md with what we completed
2. Create a commit with message: "[description of work]"
3. Tell me what we should tackle next time
```

---

## You're Ready

1. ✅ Repo created on GitHub
2. ✅ Cloned to local machine
3. ✅ Started Claude Code in the roostr directory
4. ✅ Pasted the kickoff prompt
5. 🚀 Start building!

Good luck! 🐓
