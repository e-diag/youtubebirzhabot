import { Flame } from 'lucide-react';
import { ImageWithFallback } from './figma/ImageWithFallback';

interface Listing {
  id: number;
  image: string;
  title: string;
  description: string;
  username: string;
  isPremium: boolean;
}

interface ListingCardProps {
  listing: Listing;
}

export function ListingCard({ listing }: ListingCardProps) {
  return (
    <div
      className={`bg-card rounded-2xl shadow-md overflow-hidden transition-all hover:shadow-lg ${
        listing.isPremium ? 'border-2 border-[#FF0000]' : 'border border-border'
      }`}
    >
      {/* Image */}
      <div className="relative aspect-video overflow-hidden bg-muted">
        <ImageWithFallback
          src={listing.image}
          alt={listing.title}
          className="w-full h-full object-cover"
        />
        {listing.isPremium && (
          <div className="absolute top-3 right-3 bg-[#FF0000] text-white px-3 py-1 rounded-full flex items-center gap-1 shadow-lg">
            <Flame size={16} />
            <span className="text-sm">Лучшее</span>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-4 space-y-2">
        <h3 className="line-clamp-1">{listing.title}</h3>
        <p className="text-muted-foreground text-sm line-clamp-2">{listing.description}</p>
        <a
          href={`https://t.me/${listing.username.replace('@', '')}`}
          target="_blank"
          rel="noopener noreferrer"
          className="text-[#FF0000] hover:underline inline-block"
        >
          {listing.username}
        </a>
      </div>
    </div>
  );
}
