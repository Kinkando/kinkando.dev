import {
  ShoppingCart,
  Utensils,
  Home,
  Car,
  Briefcase,
  Gift,
  HeartPulse,
  Plane,
  Wallet,
  PiggyBank,
  Coffee,
  Music,
  BookOpen,
  Gamepad2,
  Shirt,
  Dumbbell,
  Zap,
  Bus,
  TrendingUp,
  CircleDollarSign,
  type LucideIcon,
} from 'lucide-react'

export const ICON_REGISTRY: Record<string, LucideIcon> = {
  ShoppingCart,
  Utensils,
  Home,
  Car,
  Briefcase,
  Gift,
  HeartPulse,
  Plane,
  Wallet,
  PiggyBank,
  Coffee,
  Music,
  BookOpen,
  Gamepad2,
  Shirt,
  Dumbbell,
  Zap,
  Bus,
  TrendingUp,
  CircleDollarSign,
}

export const ICON_NAMES = Object.keys(ICON_REGISTRY)

export function getIcon(name: string): LucideIcon {
  return ICON_REGISTRY[name] ?? CircleDollarSign
}
