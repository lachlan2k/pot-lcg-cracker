# "Power of Two" LCG Cracker

This program can accurately predict the output of `java.util.Random`'s `nextInt(int bound)` method, when `bound` is a power of two and you have at least three samples.

If you want a more generic cracker, please use [Randcrack by fransla](https://github.com/fransla/randcrack).

## Usage

```
$ go run main.go

Usage of main:
  -bound int
        Bound value (argument passed to nextInt())
  -continue
        Find all possible matches (only do this if the output was wrong the first time)
  -gen int
        How many values to predict (default 5)
  -samples string
        List of known outputs (comma or space separated)
```

Example with sample values `53399, 58914, 24327` and a bound of `65536`. This only takes a few seconds to run.

```
$ go run main.go -bound 65536 -samples "53399, 58914, 24327"
======
Success! Found starting state 229349101478606 (1749794781<<17 + -56626)
Generating next 5 outputs:

0: 29132
1: 29127
2: 44688
3: 49749
4: 52045
======
```

These values were generated with `jshell` as seen below. You can confirm the predicted values match :).

```
jshell> Random rand = new java.util.Random();
   ...> 
rand ==> java.util.Random@52cc8049

jshell> rand.nextInt(65536);
$2 ==> 53399

jshell> rand.nextInt(65536);
$3 ==> 58914

jshell> rand.nextInt(65536);
$4 ==> 24327

jshell> rand.nextInt(65536);
$5 ==> 29132

jshell> rand.nextInt(65536);
$6 ==> 29127

jshell> rand.nextInt(65536);
$7 ==> 44688

jshell> rand.nextInt(65536);
$8 ==> 49749
```

# How does it work?

`java.util.Random` is not cryptographically secure, nor does it claim to be. There are a range of tools across the internet that allow you to predict future outputs, given a few input samples.
However, `java.util.Random` has a special case for when `bound` is a power of two that most of these tools don't cover. My tool is designed specifically for this.

## Why it sucks

The special case is seen [below](https://github.com/openjdk-mirror/jdk7u-jdk/blob/f4d80957e89a19a29bb9f9807d2a28351ed7f7df/src/share/classes/java/util/Random.java#L298):

```java
public int nextInt(int n) {
    if (n <= 0)
        throw new IllegalArgumentException("n must be positive");

    if ((n & -n) == n)  // i.e., n is a power of 2
        return (int)((n * (long)next(31)) >> 31);

    int bits, val;
    do {
        bits = next(31);
        val = bits % n;
    } while (bits - val + (n-1) < 0);
    return val;
}
```

The specific line in question is `return (int)((n * (long)next(31)) >> 31);`.

`next(n)` will take the current internal state (or "seed"), perform an LCG iteration, and output `n` pseudo-random bits, as seen [below](https://github.com/openjdk-mirror/jdk7u-jdk/blob/f4d80957e89a19a29bb9f9807d2a28351ed7f7df/src/share/classes/java/util/Random.java#L183):

```java
protected int next(int bits) {
    long oldseed, nextseed;
    AtomicLong seed = this.seed;
    do {
        oldseed = seed.get();
        nextseed = (oldseed * multiplier + addend) & mask;
    } while (!seed.compareAndSet(oldseed, nextseed));
    return (int)(nextseed >>> (48 - bits));
}
```

At the end, the final `(48 - n)` bits are dropped. In this case, that's `(48 - 31) = 17` bits. The state itself is only 48 bits in size, leaving us with 31 bits remaining.

## How we actually break it

So, given a sample value, only the first 31 bits influenced the output. Because of this, we can brute-force the first 31 bits independently of the rest. On a modern computer, `2^31` is a miniscule search space.

Then, once we've found the first 31 bits, we need more output samples to find the remaining 17 bits. So, we can simply guess from `(firstBits << 31)` to `(firsBits << 31) + 2**17` until we find a state that outputs the rest of our samples when iterated. Once we've reconstructed the state/seed, we can simply start predicting future values :).

Is there a better way of doing this? Almost certainly. But, the search space for a straight brute-force is tiny and only takes a few seconds on a modern CPU.

# Credits

[Randcrack by fransla](https://github.com/fransla/randcrack) for the inspiration and Go re-implementation of Java's LCG :)