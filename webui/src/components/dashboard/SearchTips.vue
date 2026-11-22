<template>
  <div class="search-tips">
    By default, multiple terms are combined with OR (results contain at least one term). Use prefixes to
    enforce
    stricter matches.
    <h5>Refine Your Results</h5>
    <ul>
      <li><code>+petstore -kafka</code> - Must include "petstore", Must Not include "kafka".</li>
      <li><code>petstore kafka</code> - Returns results containing "petstore" OR "kafka" (default).</li>
      <li><code>"Swagger Petstore"</code> - Matches the exact phrase.</li>
    </ul>
    <h5>Fields & Logic</h5>
    <ul>
      <li><code>name:petstore</code> - Search for "petstore" specifically in the name field.</li>
      <li><code>+method:GET 404 500</code> - Must be GET and must contain either 404 or 500.</li>
      <li><code>+statusCode:>=300</code> - Must be response with status code greater than or equal to 300.
      </li>
      <li><code>path:/pets^2</code> - Boost matches in the path field (scores them higher).</li>
    </ul>
    <h5>Wildcards & Fuzzy</h5>
    <ul>
      <li><code>pet*</code> - Wildcard (matches "pet", "pets", "petstore").</li>
      <li><code>pet~</code> - Fuzzy match (matches "pets", "pest" or slight typos).</li>
    </ul>
  </div>
</template>

<style scoped>
.search-tips {
  font-size: 0.9rem;
}

.search-tips h5 {
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  margin-top: 0.5rem;
}

.search-tips ul {
  padding-left: 1.5rem;
  margin-bottom: 0;
}
</style>
