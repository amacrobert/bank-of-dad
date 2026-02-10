import { Link } from "react-router-dom";
import { TreePine } from "lucide-react";
import Button from "../components/ui/Button";

export default function NotFound() {
  return (
    <div className="min-h-screen bg-cream flex flex-col items-center justify-center px-4">
      <div className="text-center animate-fade-in-up">
        <div className="mb-6">
          <TreePine
            className="h-20 w-20 text-sage mx-auto animate-gentle-sway"
            aria-hidden="true"
          />
        </div>
        <h1 className="text-3xl font-bold text-forest mb-3">
          Oops! This page wandered off.
        </h1>
        <p className="text-bark-light text-lg mb-8 max-w-sm mx-auto">
          The page you're looking for doesn't exist or may have been moved.
        </p>
        <Link to="/">
          <Button variant="primary">Go Home</Button>
        </Link>
      </div>
    </div>
  );
}
