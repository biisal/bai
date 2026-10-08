export default function UndoSection() {
  return (
    <section className="w-full mx-auto relative z-10">
      <div className="relative w-full min-h-[400px] md:min-h-[450px] overflow-hidden flex items-center">
        {/* Full-bleed Background Image */}
        <div className="absolute inset-0">
          <div className="absolute inset-0 bg-gradient-to-t from-transparent via-background/20 to-background z-10" />
          <div className="absolute inset-0 bg-gradient-to-l from-background/90 via-background/60 to-transparent z-10" />
          <img
            src="/undo-bg.png"
            alt="Moody coding workspace"
            className="w-full h-full object-cover"
          />
        </div>

        {/* Typography Overlay */}
        <div className="relative z-20 max-w-7xl mx-auto px-6 w-full flex justify-end">
          <div className="w-full md:w-1/3 pl-12 md:pl-0">
            <h2 className="text-4xl md:text-5xl font-['Playfair_Display'] text-white leading-[1.15] mb-4">
              Undo Anything.
              <br />
              Keep Shipping.
            </h2>

            <div className="relative inline-block self-start mt-2">
              <span className="font-['Satisfy'] text-5xl md:text-6xl text-primary -rotate-2 inline-block">
                Ship Faster
              </span>
              <svg
                className="absolute -bottom-4 left-0 w-[120%] h-6 text-primary opacity-80"
                viewBox="0 0 200 20"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                preserveAspectRatio="none"
              >
                <path
                  d="M2 18C40 10 120 -2 198 12"
                  stroke="currentColor"
                  strokeWidth="3"
                  strokeLinecap="round"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
