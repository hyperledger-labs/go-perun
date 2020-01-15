**Severity**&emsp; <!-- Choose one of the following:
	*Trivial*: Bugs that have no real impact.
	*Minor*: Minor bugs are inconvenient for users, but do not break functionality.
	*Major*: Bugs that break features or specifications.
	*Critical*: Bugs that prevent further investigation, such as crashes. -->


**Frequency**:&emsp; <!-- Choose one of the following:
	*Rare*: Almost never happens.
	*Uncommon*: Happens from time to time.
	*Common*: Happens often.
	*High*: Always happens. -->


**Location**&emsp; <!--
	Where did it happen? If not clear, give a general context.
	I.e.: [pkg/test] CheckPanic()
	Or: Peer communication -->


**Erroneus behavior**&emsp; <!--
	What happened (behaviour of the bug).
	I.e.: CheckPanic() does not detect panic(nil) calls.
	Or: Server does not respond to pings. -->


**Desired behavior**&emsp; <!-- Optional if obvious.
	What should have happened instead?
	I.e.: CheckPanic() should return a (bool, interface{}), and return whether
	panic was called, as well as the value passed to panic(). -->


**Probable cause**&emsp; <!-- Optional.
	What seemingly caused the bug or how to reproduce it.
	I.e: recover() treats panic(nil) the same as no panic.
	Or: I pressed the red button. -->


**Reproduction**&emsp; <!-- Optional if obvious.
	How can the bug be reproduced?
	Either describe or provide a short code sample. -->


**Fix**&emsp; <!-- Optional.
	How can the bug be fixed? Just a short hint would suffice. If you have no
	idea how this could be fixed, say so.
	I.e: 
		didPanic = true;
		function();
		didPanic = false; // Only executed if no panic occurred. -->


**Testing**&emsp; <!--
	Describe a test case that tests for just this bug, or, if another test case
	already covers this, name it.
	I.e: Create a sub test that tests panic(nil). -->

<!-- End -->
/label ~"🐛 bug"