# Developer Guide

## Project Purpose

Trees exists to make it extremely easy for both AI and humans to reliably gather data about code and use it as reliable research for high quality development.

Claims are statements about what code does. Evidence is a pointer to code (file + lines + git commit) that supports a claim. Evidence auto-invalidates when the referenced code changes past its recorded commit. This is the core value: you always know whether your understanding of the code is current or stale.

**This purpose should drive all design decisions.** Keep the tool's surface area small. Every flag, endpoint, and concept that gets added is another place where a human forgets something or an AI hallucinates. When in doubt, automate instead of adding options, and leave things out instead of adding them.

---

## TDD Workflow

This section outlines the Test-Driven Development (TDD) workflow used in this project, based on principles from Robert C. Martin (Uncle Bob).

## Table of Contents
1. [Why TDD Matters: Symmetry Breaking](#why-tdd-matters-symmetry-breaking)
2. [The Cycles of TDD](#the-cycles-of-tdd)
3. [How Not to Get Stuck: Transformation Priority Premise](#how-not-to-get-stuck-transformation-priority-premise)

---

## Why TDD Matters: Symmetry Breaking

**Reference:** [Symmetry Breaking](https://blog.cleancoder.com/uncle-bob/2017/03/07/SymmetryBreaking.html)

### The Importance

TDD forces us to **break symmetry** in our code. When we write tests first, we're forced to think about how code will be used before we think about how it will be implemented. This asymmetry is crucial because:

- **Prevents premature abstraction**: Without tests, we often create symmetric, generic solutions for problems we don't fully understand yet
- **Drives better design**: Tests force us to consider the API and usage patterns first
- **Reveals duplication**: As we add tests, duplicated code becomes obvious, signaling the need for abstraction
- **Creates intentional design**: Each test adds a specific requirement, breaking the symmetry of "it could work any way"

### In Practice

When you find yourself with symmetric code (duplicated logic, similar patterns), that's a signal to refactor and break the symmetry through abstraction. But only do this when the tests demand it, not preemptively.

---

## The Cycles of TDD

**Reference:** [The Cycles of TDD](https://blog.cleancoder.com/uncle-bob/2014/12/17/TheCyclesOfTDD.html)

TDD operates at multiple time scales simultaneously. Understanding these cycles helps you maintain rhythm and focus.

### The Nano-Cycle: Red-Green-Refactor (Seconds to Minutes)

This is the fundamental TDD cycle:

1. **Red**: Write a failing test
   - Think about what behavior you want
   - Write the test that specifies that behavior
   - Watch it fail (proving the test can fail)

2. **Green**: Make it pass quickly
   - Write the simplest code that makes the test pass
   - Don't worry about elegance yet
   - Hardcoding is fine at this stage

3. **Refactor**: Clean up
   - Remove duplication
   - Improve names
   - Apply design patterns where appropriate
   - Keep all tests passing

**Rhythm**: This cycle should take 30 seconds to 2 minutes. If it takes longer, your test is too big.

### The Micro-Cycle: Test by Test (Minutes)

- Add one test at a time
- Each test adds a new specific example or requirement
- Build up a suite of tests that fully specify the behavior
- Time scale: 1-10 minutes per test

### The Milli-Cycle: Feature Completion (Hours)

- Complete a full feature or user story
- Multiple micro-cycles building toward a goal
- Results in a deployable increment
- Time scale: Hours to a day

### The Primary/Secondary Cycles

- **Primary cycle**: Getting from Red to Green (writing code to pass the test)
- **Secondary cycle**: Refactoring (improving the code while keeping tests green)

### Workflow Summary

```
┌──────────────────────────────────────────────┐
│  MILLI-CYCLE: Feature/Story                  │
│  ┌────────────────────────────────────────┐  │
│  │ MICRO-CYCLE: Single Test               │  │
│  │  ┌──────────────────────────────────┐  │  │
│  │  │ NANO-CYCLE:                      │  │  │
│  │  │  RED → GREEN → REFACTOR          │  │  │
│  │  └──────────────────────────────────┘  │  │
│  │  Repeat for each test case...          │  │
│  └────────────────────────────────────────┘  │
│  ↓                                            │
│  TYPE-CHECK & LINT (make tcl)                │
│  ↓                                            │
│  COMMIT (if all checks pass)                 │
│  ↓                                            │
│  Repeat until feature complete...            │
└──────────────────────────────────────────────┘
```

---

## How Not to Get Stuck: Transformation Priority Premise

**Reference:** [The Transformation Priority Premise](https://blog.cleancoder.com/uncle-bob/2013/05/27/TheTransformationPriorityPremise.html)

### The Problem

When doing TDD, you can get stuck when the jump from the current simple code to passing the next test seems too large. You might be tempted to write too much code at once.

### The Solution: Transformations Over Refactorings

**Refactorings** change the structure of code without changing behavior.
**Transformations** change behavior by making code more general.

### The Transformation Priority List

When making a test pass, prefer transformations higher on this list (simpler) over those lower (more complex):

1. **`{} → nil`**: No code at all → code that returns nil/null
2. **`nil → constant`**: Nil/null → a simple constant value
3. **`constant → constant+`**: Simple constant → more complex constant
4. **`constant → scalar`**: Constant → variable or argument
5. **`statement → statements`**: One statement → multiple statements
6. **`unconditional → if`**: Unconditional code → conditional (if statement)
7. **`scalar → array`**: Single value → collection of values
8. **`array → container`**: Array → more complex data structure
9. **`statement → tail-recursion`**: Iteration → tail recursion
10. **`if → while`**: Conditional → loop
11. **`expression → function`**: Expression → function call/algorithm
12. **`variable → assignment`**: Replacing value of a variable

### How to Apply This

1. **When a test fails**, look at the simplest transformation that would make it pass
2. **Prefer higher priority** (simpler) transformations
3. **If you're stuck**, you probably skipped a transformation - write a simpler test that requires a higher-priority transformation
4. **Avoid big jumps** - if you need transformation #10 to pass a test, you probably need intermediate tests that use transformations #6, #7, etc.

### Example

```javascript
// Test 1: Function should exist
test('getPrime exists', () => {
  expect(getPrime).toBeDefined();
});
// Transformation: {} → nil
function getPrime() {}

// Test 2: First prime is 2
test('first prime is 2', () => {
  expect(getPrime(0)).toBe(2);
});
// Transformation: nil → constant
function getPrime() { return 2; }

// Test 3: Second prime is 3
test('second prime is 3', () => {
  expect(getPrime(1)).toBe(3);
});
// Transformation: constant → scalar, unconditional → if
function getPrime(n) {
  if (n === 0) return 2;
  return 3;
}

// Test 4: Third prime is 5
// Transformation: if → array (store primes in array)
// ... and so on
```

### Key Insight

**When you're stuck, write a simpler test.** The TPP helps you recognize when you're trying to make too big a leap and guides you toward intermediate steps.

---

## Testing "Untestable" Code: The Humble Object Pattern

**Reference:** [Humble Object Pattern](https://mercury-leo.github.io/teaching/HumbleObjectPattern/) | [Clean Coders: Advanced TDD](https://cleancoders.com/episode/clean-code-episode-23-p2)

### The Problem

Some code is inherently hard to test:
- **Browser APIs**: MediaRecorder, localStorage, geolocation
- **I/O Operations**: File system, network calls, database queries
- **UI Rendering**: DOM manipulation, canvas drawing, CSS animations
- **External Devices**: Microphone, camera, sensors

When we encounter this code, we face a choice: skip testing it (bad), or find a way to test it anyway (good).

### Uncle Bob's Rule

**"There is no such thing as untestable code."** - Uncle Bob

This is an attitude. With appropriate creativity and imagination, any code can be tested. The Humble Object pattern shows us how.

### The Solution: Separate Humble from Testable

The pattern is simple:

1. **Identify the untestable part** - the browser API, I/O operation, etc.
2. **Create a thin "humble" wrapper** - make it as simple as possible, with zero logic
3. **Put ALL logic in testable code** - state management, error handling, data transformation, business rules

The "humble" part is so simple that bugs are unlikely. All the complexity lives in testable code.

### Example: Audio Recording with Transcription

❌ **Bad: Everything mixed together (untestable)**

```typescript
function RecordButton() {
  const [isRecording, setIsRecording] = useState(false);
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);

  const handleRecord = async () => {
    if (!isRecording) {
      // Browser API call - untestable!
      const stream = await navigator.mediaDevices.getUserMedia({audio: true});
      const recorder = new MediaRecorder(stream);
      // ... lots of state management mixed with browser APIs
    }
  };
  // How do we test this without a real microphone?
}
```

✅ **Good: Humble wrapper + Testable logic**

```typescript
// 1. HUMBLE: Thin wrapper around browser API (untestable, but simple)
interface AudioRecorder {
  startRecording(): Promise<void>;
  stopRecording(): Promise<Blob>;
}

class BrowserAudioRecorder implements AudioRecorder {
  private mediaRecorder: MediaRecorder | null = null;

  async startRecording() {
    const stream = await navigator.mediaDevices.getUserMedia({audio: true});
    this.mediaRecorder = new MediaRecorder(stream);
    this.mediaRecorder.start();
  }

  async stopRecording() {
    return new Promise((resolve) => {
      this.mediaRecorder!.addEventListener('dataavailable', (e) => {
        resolve(e.data);
      });
      this.mediaRecorder!.stop();
    });
  }
}

// 2. TESTABLE: All logic separated (100% testable with mocks)
export function useAudioTranscription(recorder: AudioRecorder) {
  const [state, setState] = useState<'idle' | 'recording' | 'transcribing' | 'error'>('idle');
  const [transcript, setTranscript] = useState('');
  const [error, setError] = useState<string | null>(null);
  const client = useContext(ApiClientContext);

  const startRecording = async () => {
    try {
      setState('recording');
      await recorder.startRecording();
    } catch (err) {
      setState('error');
      setError(err.message);
    }
  };

  const stopAndTranscribe = async () => {
    try {
      const blob = await recorder.stopRecording();
      setState('transcribing');

      const formData = new FormData();
      formData.append('audio', blob, 'recording.wav');

      const response = await client.post('/ai/transcribe', formData);
      setTranscript(response.data.transcript);
      setState('idle');
    } catch (err) {
      setState('error');
      setError(err.message);
    }
  };

  return {state, transcript, error, startRecording, stopAndTranscribe};
}

// 3. TESTS: Test the logic with a mock recorder
describe('useAudioTranscription', () => {
  it('transitions to recording state when startRecording is called', async () => {
    const mockRecorder = {
      startRecording: vi.fn().mockResolvedValue(undefined),
      stopRecording: vi.fn(),
    };

    const {result} = renderHook(() => useAudioTranscription(mockRecorder));

    await act(() => result.current.startRecording());

    expect(result.current.state).toBe('recording');
    expect(mockRecorder.startRecording).toHaveBeenCalled();
  });

  it('posts audio blob to API and returns transcript', async () => {
    const mockBlob = new Blob(['fake audio'], {type: 'audio/wav'});
    const mockRecorder = {
      startRecording: vi.fn(),
      stopRecording: vi.fn().mockResolvedValue(mockBlob),
    };
    const mockPost = vi.fn().mockResolvedValue({
      data: {transcript: 'Hello world'}
    });

    const {result} = renderHook(() => useAudioTranscription(mockRecorder), {
      wrapper: ({children}) => (
        <ApiClientContext.Provider value={{post: mockPost}}>
          {children}
        </ApiClientContext.Provider>
      ),
    });

    await act(() => result.current.stopAndTranscribe());

    expect(mockPost).toHaveBeenCalledWith('/ai/transcribe', expect.any(FormData));
    expect(result.current.transcript).toBe('Hello world');
    expect(result.current.state).toBe('idle');
  });
});
```

### Key Benefits

1. **High test coverage**: 95%+ of the code is tested, only the thin wrapper is untested
2. **Fast tests**: No real I/O, runs in milliseconds
3. **Reliable tests**: No flaky tests from timing issues or external dependencies
4. **Easy mocking**: Simple interface makes mocking trivial
5. **Maintainable**: Clear separation makes code easier to understand and change

### When to Apply This Pattern

Use the Humble Object pattern whenever you encounter:
- Browser APIs (MediaRecorder, localStorage, geolocation, etc.)
- Network requests (already abstracted via ApiClientContext)
- File system operations
- External hardware (microphone, camera, GPS)
- UI framework internals (React rendering, animation frames)
- Date/time (use dependency injection for Date.now())

### Guidelines

1. **Make the humble part as dumb as possible** - no conditionals, no loops, no calculations
2. **Test the testable part thoroughly** - aim for 100% coverage
3. **Use dependency injection** - pass the humble object to the testable code
4. **Accept that some code is untested** - but minimize it to ~5% of total code
5. **Document why code is humble** - add comments explaining the untestable dependency

---

## Quick Reference

### Daily TDD Checklist

- [ ] Write a failing test (Red)
- [ ] Make it pass with simplest code (Green)
- [ ] Refactor while keeping tests green
- [ ] **Run type-check and lint** (`make tcl`)
- [ ] Commit when all tests pass AND type-check/lint pass
- [ ] Prefer simpler transformations over complex ones
- [ ] If stuck, write a simpler test
- [ ] Break symmetry through abstraction only when tests demand it

### Before Every Commit

Run these commands to ensure code quality:

```bash
make tcl    # Run type-check and lint (web-app)
make test   # Run impacted tests
```

Or run the full test suite:

```bash
make test-a  # Run all tests (web-app + backend)
```

**Never commit code that fails type-check or lint.** These checks catch bugs early and ensure code consistency.

### When You're Stuck

1. Is your test too big? Break it into smaller tests
2. What's the simplest transformation that would help?
3. Are you trying to jump too far down the transformation list?
4. Do you need an intermediate test?

---

## Additional Resources

- [Symmetry Breaking](https://blog.cleancoder.com/uncle-bob/2017/03/07/SymmetryBreaking.html)
- [The Cycles of TDD](https://blog.cleancoder.com/uncle-bob/2014/12/17/TheCyclesOfTDD.html)
- [Transformation Priority Premise](https://blog.cleancoder.com/uncle-bob/2013/05/27/TheTransformationPriorityPremise.html)
