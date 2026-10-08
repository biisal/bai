import { GitBranch, Database, Palette, BookOpen, Bell } from "lucide-react";

export default function FeatureBar() {
  const features = [
    {
      icon: <GitBranch className="w-5 h-5" />,
      title: "Shadow Git",
      desc: "Invisible undo history",
    },
    {
      icon: <Database className="w-5 h-5" />,
      title: "Sessions",
      desc: "Local & resumable",
    },
    {
      icon: <Palette className="w-5 h-5" />,
      title: "Themes",
      desc: "Hot-swappable colors",
    },
    {
      icon: <BookOpen className="w-5 h-5" />,
      title: "Skills",
      desc: "Global & local SKILL.md",
    },
    {
      icon: <Bell className="w-5 h-5" />,
      title: "Notifications",
      desc: "Herdr & audio alerts",
    },
  ];

  return (
    <section className="w-full max-w-7xl mx-auto px-6 -mt-12 relative z-30">
      <div className="rounded-sm bg-muted backdrop-blur-md border border-border p-6 shadow-2xl overflow-x-auto no-scrollbar">
        <div className="flex items-center justify-between min-w-[900px] divide-x divide-border">
          {features.map((item, index) => (
            <div
              key={index}
              className="flex-1 flex items-center gap-4 px-6 hover:bg-white/[0.02] rounded-sm transition-colors py-2 cursor-default group"
            >
              <div className="w-12 h-12 rounded-sm border border-border flex items-center justify-center bg-muted text-muted-foreground group-hover:text-primary group-hover:border-primary/30 transition-all">
                {item.icon}
              </div>
              <div className="flex flex-col">
                <span className="text-white font-medium text-sm">
                  {item.title}
                </span>
                <span className="text-muted-foreground text-xs">
                  {item.desc}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
