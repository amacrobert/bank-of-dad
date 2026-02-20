import { useState } from "react";
import { Child } from "../types";
import Card from "./ui/Card";
import ChildList from "./ChildList";
import AddChildForm from "./AddChildForm";
import ChildAccountSettings from "./ChildAccountSettings";

export default function ChildrenSettings() {
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [selectedChild, setSelectedChild] = useState<Child | null>(null);

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildDeleted = () => {
    setSelectedChild(null);
    setChildRefreshKey((k) => k + 1);
  };

  return (
    <div className="md:grid md:grid-cols-[300px_1fr] md:gap-6">
      {/* Left column: child list + add child */}
      <div className="space-y-4 mb-6 md:mb-0">
        <Card padding="md">
          <ChildList
            refreshKey={childRefreshKey}
            onSelectChild={setSelectedChild}
            selectedChildId={selectedChild?.id}
          />
        </Card>

        <AddChildForm onChildAdded={handleChildAdded} />
      </div>

      {/* Right column: account settings for selected child */}
      <div>
        {selectedChild ? (
          <ChildAccountSettings
            key={selectedChild.id}
            child={selectedChild}
            onUpdated={handleChildUpdated}
            onDeleted={handleChildDeleted}
          />
        ) : (
          <Card padding="lg" className="hidden md:flex items-center justify-center min-h-[300px]">
            <p className="text-bark-light text-center">
              Select a child to manage their account settings.
            </p>
          </Card>
        )}
      </div>
    </div>
  );
}
