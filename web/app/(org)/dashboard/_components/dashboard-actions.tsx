"use client";

import {
  Briefcase,
  Factory,
  FileDown,
  FileText,
  Package,
  Plus,
  ShoppingCart,
  Truck,
  UserRound,
  Users,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/dropdown";
import { toast } from "@/components/toast";

export function DashboardActions() {
  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => toast("Report export has been queued. You'll get an email when it's ready.")}
      >
        <FileDown aria-hidden />
        <span className="hidden sm:inline">{"Export report"}</span>
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button size="sm">
            <Plus aria-hidden />
            <span className="hidden sm:inline">{"New record"}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem render={<Link href={"/contacts"} />}>
            <Users aria-hidden />
            {"New customer"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/purchases/suppliers"} />}>
            <Truck aria-hidden />
            {"New supplier"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/products"} />}>
            <Package aria-hidden />
            {"New item"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/sale-orders/customers"} />}>
            <ShoppingCart aria-hidden />
            {"New sale order"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/purchases/suppliers"} />}>
            <Truck aria-hidden />
            {"New purchase order"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/accounting"} />}>
            <FileText aria-hidden />
            {"New invoice"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/employees"} />}>
            <UserRound aria-hidden />
            {"New employee"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/products/recipes"} />}>
            <Factory aria-hidden />
            {"New manufacturing order"}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/dashboard"} />}>
            <Briefcase aria-hidden />
            {"New project"}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
