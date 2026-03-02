import {
  extremeList,
  savageList,
  criterionList,
  ultimateList,
} from "@/config/constants";

import FightGuideComponent from "./FightGuideComponent";

export default function EncounterEntry() {
  const ultimateEntries = ultimateList.slice(0, 1);
  const savageEntries = savageList.slice(0, 4);
  const criterionEntries = criterionList.slice(0, 1);
  const extremeEntries = extremeList.slice(0, 1);

  return (
    <div>
      {ultimateEntries.length > 0 && (
        <FightGuideComponent
          entries={ultimateEntries}
          title="Current Ultimate"
          left
        />
      )}
      <FightGuideComponent
        entries={savageEntries}
        title="Current Savage Tier"
      />
      {criterionEntries.length > 0 && (
        <FightGuideComponent
          entries={criterionEntries}
          title="Current Criterion"
        />
      )}
      {extremeEntries.length > 0 && (
        <FightGuideComponent
          entries={extremeEntries}
          title="Current Extreme Trial"
        />
      )}
    </div>
  );
}
