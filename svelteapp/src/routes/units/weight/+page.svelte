<script>
  //TODO
  // move descriptions to db
  // its weird to hardcore them
  // now it's just draft code
  //--
  // how to generate units pages automatically?
  // it's will became difficult to support by hands tonns of units
  const units = [
    {
      name: "metric",
      list: [
        "gram",
        "kilogram",
        "kiloton",
        "ton",
        "centner",
        "centigram",
        "milligram",
        "microgram",
        "megaton",
        "gigatonne",
        "dekagram",
        "decigram",
        "hectogram",
      ],
    },
    {
      name: "troy",
      list: ["pound", "ounce", "pennyweight", "carat", "grain", "mite", "doit"],
    },
  ];
  units.map((e) => {
    e.list.sort();
  });
</script>

<div class="wrapper">
  <h1>Groups</h1>
  <div class="groups">
    {#each units as group}
      <h2>{group.name}</h2>
      <div class="units-list">
        {#each group.list as unit}
          <div class="unit">
            <p>{unit}</p>
          </div>
        {/each}
      </div>
    {/each}
  </div>
</div>

<style>
    .units-list {
        display: grid;
        grid-template-columns: repeat(6, 1fr);
        column-gap: 12px;
        row-gap: 12px;
    }
    .unit {
        padding: 12px;
        border: 1px solid orange;
    }
</style>