import { useState, FormEvent } from "react";
import Input from "../components/ui/Input";
import Select from "../components/ui/Select";
import Button from "../components/ui/Button";
import { CreateChoreRequest } from "../api";
import { Chore, ChoreRecurrence } from "../types";

interface ChoreFormProps {
  children: { id: number; first_name: string }[];
  onSubmit: (data: CreateChoreRequest) => Promise<void>;
  onCancel: () => void;
  loading: boolean;
  editChore?: Chore;
}

export default function ChoreForm({ children, onSubmit, onCancel, loading, editChore }: ChoreFormProps) {
  const [name, setName] = useState(editChore?.name ?? "");
  const [description, setDescription] = useState(editChore?.description ?? "");
  const [rewardDollars, setRewardDollars] = useState(
    editChore ? (editChore.reward_cents / 100).toFixed(2) : ""
  );
  const [recurrence, setRecurrence] = useState<ChoreRecurrence>(
    (editChore?.recurrence as ChoreRecurrence) ?? "one_time"
  );
  const [dayOfWeek, setDayOfWeek] = useState<number>(editChore?.day_of_week ?? 0);
  const [dayOfMonth, setDayOfMonth] = useState<number>(editChore?.day_of_month ?? 1);
  const [selectedChildIds, setSelectedChildIds] = useState<number[]>(
    editChore?.assignments?.map((a) => a.child_id) ?? []
  );
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = "Name is required";
    } else if (name.trim().length > 100) {
      newErrors.name = "Name must be 100 characters or less";
    }

    if (description.length > 500) {
      newErrors.description = "Description must be 500 characters or less";
    }

    const reward = parseFloat(rewardDollars);
    if (!rewardDollars || isNaN(reward) || reward < 0) {
      newErrors.reward = "Reward must be a positive amount";
    }

    if (recurrence === "monthly") {
      if (dayOfMonth < 1 || dayOfMonth > 31) {
        newErrors.dayOfMonth = "Day must be between 1 and 31";
      }
    }

    if (selectedChildIds.length === 0) {
      newErrors.children = "At least one child must be selected";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    const rewardCents = Math.round(parseFloat(rewardDollars) * 100);

    const data: CreateChoreRequest = {
      name: name.trim(),
      reward_cents: rewardCents,
      recurrence,
      child_ids: selectedChildIds,
    };

    if (description.trim()) {
      data.description = description.trim();
    }

    if (recurrence === "weekly") {
      data.day_of_week = dayOfWeek;
    }

    if (recurrence === "monthly") {
      data.day_of_month = dayOfMonth;
    }

    await onSubmit(data);
  };

  const toggleChild = (childId: number) => {
    setSelectedChildIds((prev) =>
      prev.includes(childId)
        ? prev.filter((id) => id !== childId)
        : [...prev, childId]
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h2 className="text-lg font-bold text-bark">{editChore ? "Edit Chore" : "New Chore"}</h2>

      <Input
        label="Name"
        id="chore-name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        maxLength={100}
        placeholder="e.g., Take out the trash"
        error={errors.name}
        required
      />

      <div className="space-y-1.5">
        <label htmlFor="chore-description" className="block text-sm font-semibold text-bark-light">
          Description (optional)
        </label>
        <textarea
          id="chore-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          maxLength={500}
          placeholder="Add details about how to complete this chore..."
          rows={3}
          className={`
            w-full px-4 py-3
            rounded-xl border border-sand bg-white
            text-bark text-base placeholder:text-bark-light/50
            transition-all duration-200
            focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
            ${errors.description ? "border-terracotta ring-1 ring-terracotta/30" : ""}
          `}
        />
        {errors.description && (
          <p className="text-sm text-terracotta font-medium">{errors.description}</p>
        )}
      </div>

      <Input
        label="Reward ($)"
        id="chore-reward"
        type="number"
        step="0.01"
        min="0"
        value={rewardDollars}
        onChange={(e) => setRewardDollars(e.target.value)}
        placeholder="5.00"
        error={errors.reward}
        required
      />

      <Select
        label="Recurrence"
        id="chore-recurrence"
        value={recurrence}
        onChange={(e) => setRecurrence(e.target.value as ChoreRecurrence)}
      >
        <option value="one_time">One-time</option>
        <option value="daily">Daily</option>
        <option value="weekly">Weekly</option>
        <option value="monthly">Monthly</option>
      </Select>

      {recurrence === "weekly" && (
        <Select
          label="Day of Week"
          id="chore-day-of-week"
          value={dayOfWeek}
          onChange={(e) => setDayOfWeek(parseInt(e.target.value))}
        >
          <option value={0}>Sunday</option>
          <option value={1}>Monday</option>
          <option value={2}>Tuesday</option>
          <option value={3}>Wednesday</option>
          <option value={4}>Thursday</option>
          <option value={5}>Friday</option>
          <option value={6}>Saturday</option>
        </Select>
      )}

      {recurrence === "monthly" && (
        <Input
          label="Day of Month"
          id="chore-day-of-month"
          type="number"
          min={1}
          max={31}
          value={dayOfMonth}
          onChange={(e) => setDayOfMonth(parseInt(e.target.value))}
          error={errors.dayOfMonth}
        />
      )}

      <div className="space-y-1.5">
        <span className="block text-sm font-semibold text-bark-light">Assign to</span>
        <div className="space-y-2">
          {children.map((child) => (
            <label
              key={child.id}
              className="flex items-center gap-3 cursor-pointer"
            >
              <input
                type="checkbox"
                checked={selectedChildIds.includes(child.id)}
                onChange={() => toggleChild(child.id)}
                className="h-4 w-4 rounded accent-forest"
              />
              <span className="text-sm text-bark font-medium">{child.first_name}</span>
            </label>
          ))}
        </div>
        {errors.children && (
          <p className="text-sm text-terracotta font-medium">{errors.children}</p>
        )}
      </div>

      <div className="flex gap-3 pt-2">
        <Button type="submit" loading={loading} className="flex-1">
          {editChore ? "Save Changes" : "Create Chore"}
        </Button>
        <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
