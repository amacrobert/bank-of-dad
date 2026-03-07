import { Leaf } from "lucide-react";

interface FooterProps {
  variant?: "dark" | "subtle";
}

export default function Footer({ variant = "dark" }: FooterProps) {
  const dark = variant === "dark";
  return (
    <footer className={dark ? "bg-bark" : ""}>
      <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8 flex flex-col gap-6">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Leaf className={`h-5 w-5 ${dark ? "text-sage" : "text-sage/50"}`} aria-hidden="true" />
            <span className={`text-sm font-bold ${dark ? "text-white" : "text-bark-light/50"}`}>Bank of Dad</span>
          </div>
          <nav className="flex items-center gap-6">
            {/* Future links go here */}
          </nav>
        </div>
        <div className={`border-t ${dark ? "border-white/10" : "border-sand"} pt-4`}>
          <p className={`text-sm ${dark ? "text-sage-light/70" : "text-bark-light/40"} text-center sm:text-left`}>
            &copy; {new Date().getFullYear()} Bank of Dad. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  );
}
