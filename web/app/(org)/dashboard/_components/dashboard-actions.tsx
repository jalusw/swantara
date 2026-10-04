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
import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/dropdown";
import { toast } from "@/components/toast";

export function DashboardActions() {
  const t = useTranslations("Dashboard");
  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={() => toast(t("reportExportQueued"))}>
        <FileDown aria-hidden />
        <span className="hidden sm:inline">{t("exportReport")}</span>
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button size="sm">
            <Plus aria-hidden />
            <span className="hidden sm:inline">{t("quickCreate")}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem render={<Link href={"/contacts"} />}>
            <Users aria-hidden />
            {t("newCustomer")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/purchases/suppliers"} />}>
            <Truck aria-hidden />
            {t("newSupplier")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/products"} />}>
            <Package aria-hidden />
            {t("newProduct")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/sale-orders/customers"} />}>
            <ShoppingCart aria-hidden />
            {t("newSalesOrder")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/purchases/suppliers"} />}>
            <Truck aria-hidden />
            {t("newPurchaseOrder")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/accounting"} />}>
            <FileText aria-hidden />
            {t("newInvoice")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/employees"} />}>
            <UserRound aria-hidden />
            {t("newEmployee")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/products/recipes"} />}>
            <Factory aria-hidden />
            {t("newProductionOrder")}
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link href={"/dashboard"} />}>
            <Briefcase aria-hidden />
            {t("newProject")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
