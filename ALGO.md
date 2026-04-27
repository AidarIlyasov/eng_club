# English Club Session Manager - Algorithm Explanation

## Problem Overview

The English Club needs to organize sessions with three activities:
1. **One-to-One Talks** (15 minutes) - Members pair up for 2-person conversations
2. **Small Group Discussions** (30 minutes) - Members form groups of 3 for discussions
3. **Alias Game** (15 minutes) - Members form groups of 3 for playing games

**Key Constraint**: No two members should be in the same group more than once during a single session.

## Algorithm Design

### Core Approach: Constrained Greedy Matching with History Tracking

The solution uses two main algorithms:

### 1. Weighted Greedy Matching (for One-to-One Pairs)

**Purpose**: Create optimal pairs while minimizing repeated pairings across sessions.

**Algorithm Steps**:
1. **Shuffle** the member list for randomization
2. **Greedy Selection**: For each unmatched member:
   - Find the best available partner
   - "Best" = partner with lowest historical pairing count
   - This ensures variety across multiple sessions
3. **Constraint Enforcement**: Track all pairs to prevent duplicates in later activities

**Time Complexity**: O(n²) where n = number of members

**Example**:
```
Members: [1, 2, 3, 4]
Shuffle: [3, 1, 4, 2]

Iteration 1: Member 3 pairs with Member 1 (weight=0, first time together)
Iteration 2: Member 4 pairs with Member 2 (weight=0, first time together)
Result: [(3,1), (4,2)]
```

### 2. Constrained Greedy Grouping (for Small Groups and Alias Games)

**Purpose**: Form groups of 3 while avoiding pairs that already occurred in the session.

**Algorithm Steps**:
1. **Shuffle** for randomization
2. **Build Groups Iteratively**:
   - Start with an empty group
   - For each unassigned member, check if they can join without creating duplicate pairs
   - Add compatible members until group size reaches 3
3. **Fallback Strategy**: If constraints are too strict (rare edge case):
   - Relax constraints and take any available members
   - This ensures all members get assigned

**Time Complexity**: O(n² × g) where g = number of groups

**Example**:
```
Members: [1, 2, 3, 4, 5, 6]
Already paired in session: [(1,2), (3,4), (5,6)]

Group 1: Try [1, 3, 5] - Valid! (no existing pairs)
Group 2: Try [2, 4, 6] - Valid! (no existing pairs)
```

## Data Structures

### 1. Pair Tracking
- **sessionPairs**: `map[string]bool` - Tracks all pairs within current session
- **PairHistory**: `map[string]int` - Counts pair occurrences across all sessions
- **Key Format**: "memberID1-memberID2" (smaller ID first for consistency)

### 2. Member Management
- **Member struct**: Stores ID and name
- **Session struct**: Contains all activities for one session
- **ClubManager**: Orchestrates sessions and maintains history

## Constraint Satisfaction

The algorithm ensures:

### Within-Session Constraints
- **Hard Constraint**: No pair appears together more than once per session
- **Implementation**: `sessionPairs` map prevents duplicate pairings
- **Validation**: Before adding any pairing, check `sessionPairs[key]`

### Cross-Session Optimization
- **Soft Constraint**: Minimize repeated pairings across sessions
- **Implementation**: Weight-based selection using `PairHistory`
- **Result**: More variety and better social mixing over time

## Edge Cases Handled

1. **Odd Number of Members**:
   - One-to-one: One member remains unpaired
   - Groups: One group may have 2 or 4 members

2. **Constraint Conflicts**:
   - Fallback mechanism ensures all members get assigned
   - Prioritizes participation over perfect constraint satisfaction

3. **Variable Attendance**:
   - System adapts to different member counts each session
   - History tracking maintains continuity

## Performance Characteristics

- **Space Complexity**: O(n²) for pair tracking
- **Time Complexity**: O(n²) per session
- **Scalability**: Handles up to ~100 members efficiently

## Real-World Benefits

1. **Maximizes Social Interaction**: Everyone meets different people
2. **Fair Distribution**: No one gets stuck with same partners
3. **Flexible**: Adapts to varying attendance
4. **Historical Awareness**: Improves variety over multiple sessions

## Example Output

```
Session 1 (10 members):
- One-to-One: 5 unique pairs
- Small Groups: ~3 groups of 3 (no repeated pairs)
- Alias Games: ~3 groups of 3 (no repeated pairs)

Session 2 (14 members):
- Algorithm remembers Session 1 pairings
- Prioritizes new pairings over repeated ones
- Still ensures no duplicates within Session 2
```

This design balances mathematical constraints with practical club needs, creating an engaging experience where members consistently meet new people while maintaining session structure.