<!--
	The feature template is used for features that are to be implemented. Here,
	you can work on the concept for a merge request.
	If you are not sure whether this will actually be implemented, better open
	a SUGGESTION issue first, where you can discuss suggested features.
-->

**Location**&emsp; <!--
	Package name and feature name.
	I.e: [wire] Serialization code -->


**Discussed in**&emsp; <!--
	Only if there exists one.
	The suggestion issue number that was used to discuss this feature or a link
	to a discussion in another issue or merge request.
	I.e: #66 -->


**Description**&emsp; <!--
	Describe the feature. Reminder:
	* What does the feature do (i.e., give interface)?
	* How is it done (high level information that is important)?
	* What are the acceptance criteria (i.e., what tests are needed)?
	  This can be ignored if it is obvious.
	I.e: Add support for []byte values in wire.Encode() and wire.Decode(). Do
	not send the slice's length, and when receiving, you need to know the slice
	length beforehand. To receive a slice, you need to allocate it to the right
	size beforehand. -->


**Context**&emsp; <!--
	Why / where is the feature needed? This gives the issue more context.
	I.e: We need byte slice serialization for big.Ints and for the channel.ID. -->


**Thread safety**&emsp; <!--
	Does the component need to be thread safe?
	Yes: Which functions need to be thread safe?
	No: Why not? -->


**Testing**&emsp; <!--
	What should be tested?
	Where is the test data coming from?
	Are there existing Generic... tests that should be used?
	How is the thread safety (if any) tested? -->


**Acceptance criteria**&emsp; <!--
	What does the reviewer need to check?
	When can the reviewer accept the PR? -->


**Suggested implementer and reviewers**&emsp; <!--
	Who should implement this?
	Who is reviewing the concept?
	Who is reviewing the implementation? -->


**Implementation hints**&emsp; <!--
	Hints on how to realize the implementation (useful if someone else has to do it).
	I.e: Use Writer.Write() and io.ReadFull(). -->


<!-- End -->
/label ~"💡concept"
