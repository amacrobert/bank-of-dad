import { useState } from "react";
import { post, ApiRequestError } from "../api";
import { ChildCreateResponse } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import AvatarPicker from "./AvatarPicker";
import { CheckCircle } from "lucide-react";

interface AddChildFormProps {
  onChildAdded: () => void;
}

export default function AddChildForm({ onChildAdded }: AddChildFormProps) {
  const [firstName, setFirstName] = useState("");
  const [password, setPassword] = useState("");
  const [avatar, setAvatar] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<ChildCreateResponse | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setCreated(null);

    try {
      const result = await post<ChildCreateResponse>("/children", {
        first_name: firstName,
        password,
        avatar: avatar || undefined,
      });
      setCreated(result);
      setFirstName("");
      setPassword("");
      setAvatar(null);
      onChildAdded();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card padding="md">
      <h3 className="text-base font-bold text-bark mb-3">Add a new child</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="First Name"
          id="child-name"
          type="text"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
          required
          disabled={submitting}
        />
        <Input
          label="Password (min 6 characters)"
          id="child-password"
          type="text"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          minLength={6}
          required
          disabled={submitting}
        />

        <AvatarPicker selected={avatar} onSelect={setAvatar} />

        {error && (
          <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
            <p className="text-sm text-terracotta font-medium">{error}</p>
          </div>
        )}

        <Button type="submit" loading={submitting} className="w-full">
          {submitting ? "Creating..." : "Add Child"}
        </Button>
      </form>

      {created && (
        <div className="mt-4 bg-forest/5 border border-forest/15 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <CheckCircle className="h-5 w-5 text-forest" aria-hidden="true" />
            <span className="font-bold text-forest">Account created for {created.first_name}!</span>
          </div>
          <div className="space-y-1 text-sm text-bark-light">
            <p>Login URL: <strong className="text-bark">{window.location.host}/{created.login_url}</strong></p>
            <p>Name: <strong className="text-bark">{created.first_name}</strong></p>
            <p>Password: <strong className="text-bark">{password || "(the password you just set)"}</strong></p>
          </div>
          <p className="mt-2 text-xs text-bark-light">Share these credentials with your child.</p>
        </div>
      )}
    </Card>
  );
}
