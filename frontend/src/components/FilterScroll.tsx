import { useRef } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from './ui/button';

interface FilterScrollProps {
  children: React.ReactNode;
}

export function FilterScroll({ children }: FilterScrollProps) {
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const scroll = (direction: 'left' | 'right') => {
    if (scrollContainerRef.current) {
      const scrollAmount = 200;
      const newScrollLeft = scrollContainerRef.current.scrollLeft + 
        (direction === 'left' ? -scrollAmount : scrollAmount);
      
      scrollContainerRef.current.scrollTo({
        left: newScrollLeft,
        behavior: 'smooth'
      });
    }
  };

  return (
    <div className="relative flex items-center gap-2">
      <Button
        onClick={() => scroll('left')}
        variant="ghost"
        size="icon"
        className="flex-shrink-0 h-8 w-8 rounded-full hover:bg-muted"
      >
        <ChevronLeft size={20} />
      </Button>
      
      <div
        ref={scrollContainerRef}
        className="flex gap-2 overflow-x-auto scrollbar-hide pb-2 flex-1"
        style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
      >
        {children}
      </div>
      
      <Button
        onClick={() => scroll('right')}
        variant="ghost"
        size="icon"
        className="flex-shrink-0 h-8 w-8 rounded-full hover:bg-muted"
      >
        <ChevronRight size={20} />
      </Button>
    </div>
  );
}
