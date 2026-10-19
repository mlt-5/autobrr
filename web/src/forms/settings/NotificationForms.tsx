/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { XMarkIcon } from "@heroicons/react/24/solid";
import Select from "react-select";
import { useSelector } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { APIClient } from "@api/APIClient";
import { NotificationKeys } from "@api/query_keys";
import { ExternalFilterWebhookMethodOptions, getEventOptions, getNotificationTypeOptions, SelectOption } from "@domain/constants";
import { useAppForm, useFormContext, useFormValue, fieldErrors } from "@hooks/form";
import type { FormFieldErrors } from "@hooks/form";
import { FormDebug } from "@components/debug";
import { SlideOver, SlideOverShell, SlideOverTitle } from "@components/panels";
import { toast } from "@components/hot-toast";
import Toast from "@components/notifications/Toast";
import { selectComponents, selectStyles, selectTheme } from "@components/inputs/select_props";
import { PasswordFieldWide, SelectFieldWide, SwitchGroupWide, TextFieldWide } from "@components/inputs";
import { ErrorField } from "@components/inputs/common";
import { Checkbox } from "@components/Checkbox";
import { EmptySimple } from "@components/emptystates";
import { browserNotificationsSupported } from "@hooks/useInbox";
import { SettingsContext } from "@utils/Context";

import { componentMapType } from "./types";
import { AddFormProps, UpdateFormProps } from "@forms/_shared";

type NotificationEventOption = ReturnType<typeof getEventOptions>[number];

function FormFieldsGenericWebhook() {
  const { t } = useTranslation("settings");
  return (
    <div className="border-t border-gray-200 dark:border-gray-700 py-4">
      <div className="px-4">
        <h2 className="text-lg font-medium text-gray-900 dark:text-white">
          {t("forms.notification.settings")}
        </h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {t("forms.notification.settingsDescWebhook")}
        </p>
      </div>

      <PasswordFieldWide
        name="webhook"
        label={t("forms.notification.webhookUrl")}
        help={t("forms.notification.webhookUrl")}
        placeholder={t("forms.notification.webhookPlaceholder")}
        required={true}
      />
      <SelectFieldWide
        name="method"
        label={t("forms.notification.httpMethod")}
        optionDefaultText={t("forms.notification.httpMethodDefault")}
        options={ExternalFilterWebhookMethodOptions}
        tooltip={<p>{t("forms.notification.httpMethodTooltip")}</p>}
      />
      <TextFieldWide
        name="headers"
        label={t("forms.notification.customHeaders")}
        help={t("forms.notification.customHeadersHelp")}
        placeholder={t("forms.notification.customHeadersPlaceholder")}
      />
    </div>
  );
}

function FormFieldsBuiltin() {
  const { t } = useTranslation("settings");
  const browserNotifications = SettingsContext.useSelector((s) => s.browserNotifications);
  const supported = browserNotificationsSupported() && window.isSecureContext;

  const setBrowserNotifications = async (enabled: boolean) => {
    if (enabled && Notification.permission !== "granted" && await Notification.requestPermission() !== "granted") {
      toast.custom((toastInstance) => <Toast type="warning" body={t("forms.notification.browserNotificationsDenied")} t={toastInstance} />);
      return;
    }

    SettingsContext.set((prev) => ({ ...prev, browserNotifications: enabled }));
  };

  return (
    <div className="border-t border-gray-200 dark:border-gray-700 py-4">
      <div className="px-4">
        <h2 className="text-lg font-medium text-gray-900 dark:text-white">
          {t("forms.notification.settings")}
        </h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {t("forms.notification.settingsDescBuiltin")}
        </p>
      </div>

      <Checkbox
        name="browser_notifications"
        label={t("forms.notification.browserNotifications")}
        description={supported
          ? t("forms.notification.browserNotificationsHelp")
          : t("forms.notification.browserNotificationsUnsupported")}
        value={supported && browserNotifications}
        setValue={(enabled) => void setBrowserNotifications(enabled)}
        disabled={!supported}
        className="p-4"
      />
    </div>
  );
}

const componentMap: componentMapType = {
  WEBHOOK: <FormFieldsGenericWebhook />,
  BUILTIN: <FormFieldsBuiltin />
};

interface NotificationFormValues {
  name: string;
  type: NotificationType | "";
  webhook?: string;
}

const validateNotification = (values: NotificationFormValues, required: string) => {
  const errors: FormFieldErrors = {};
  if (!values.name)
    errors.name = required;
  if (!values.type)
    errors.type = required;
  if (values.type === "WEBHOOK" && !values.webhook?.trim())
    errors.webhook = required;

  return errors;
};

export function NotificationAddForm({ isOpen, toggle }: AddFormProps) {
  return (
    <SlideOverShell isOpen={isOpen} toggle={toggle}>
      <NotificationAddFormPanel toggle={toggle} />
    </SlideOverShell>
  );
}

interface NotificationAddFormPanelProps {
  toggle: () => void;
}

function NotificationAddFormPanel({ toggle }: NotificationAddFormPanelProps) {
  const { t } = useTranslation(["options", "settings"]);
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (notification: ServiceNotification) => APIClient.notifications.create(notification),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NotificationKeys.lists() });

      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.notification.added")} t={toastInstance} />);
      toggle();
    },
    onError: () => {
      toast.custom((toastInstance) => <Toast type="error" body={t("settings:forms.notification.addFailed")} t={toastInstance} />);
    }
  });

  const onSubmit = (formData: unknown) => createMutation.mutate(formData as ServiceNotification);

  const testMutation = useMutation({
    mutationFn: (n: ServiceNotification) => APIClient.notifications.test(n)
  });

  const testNotification = (data: unknown) => testMutation.mutate(data as ServiceNotification);

  const validate = (values: NotificationFormValues) => validateNotification(values, t("settings:forms.notification.required"));

  const initialValues = {
    enabled: true,
    type: "" as NotificationType | "",
    name: "",
    webhook: "",
    events: []
  };

  const form = useAppForm({
    defaultValues: initialValues,
    validators: {
      onChange: ({ value }) => fieldErrors(validate(value))
    },
    onSubmit: ({ value }) => onSubmit(value)
  });

  const type = useSelector(form.store, (state) => state.values.type);

  return (
    <form.AppForm>
      <form
        className="h-full min-h-0 flex flex-col bg-white dark:bg-gray-800"
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="px-4 py-6 bg-gray-50 dark:bg-gray-900 sm:px-6">
            <div className="flex items-start justify-between space-x-3">
              <div className="space-y-1">
                <SlideOverTitle>
                  {t("settings:forms.notification.addTitle")}
                </SlideOverTitle>
                <p className="text-sm text-gray-500 dark:text-gray-200">
                  {t("settings:forms.notification.addDescription")}
                </p>
              </div>
              <div className="h-7 flex items-center">
                <button
                  type="button"
                  className="bg-white dark:bg-gray-700 rounded-md text-gray-400 hover:text-gray-500 focus:outline-hidden focus:ring-2 focus:ring-blue-500 cursor-pointer"
                  onClick={toggle}
                >
                  <span className="sr-only">{t("settings:forms.notification.closePanel")}</span>
                  <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>

          <div className="flex flex-col space-y-4 px-1 pt-6 sm:py-0 sm:space-y-0">
            <TextFieldWide
              name="name"
              label={t("settings:forms.notification.name")}
              required={true}
            />

            <div className="flex items-center justify-between space-y-1 px-4 sm:space-y-0 sm:grid sm:grid-cols-3 sm:gap-4">
              <div>
                <label
                  htmlFor="type"
                  className="block text-sm font-medium text-gray-900 dark:text-white"
                >
                  {t("settings:forms.notification.type")}
                </label>
              </div>
              <div className="sm:col-span-2">
                <NotificationTypeSelector />
              </div>
            </div>

            <SwitchGroupWide name="enabled" label={t("settings:forms.notification.enabled")} />

            <div className="border-t border-gray-200 dark:border-gray-700 py-4">
              <div className="px-4">
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                  {t("settings:forms.notification.globalEvents")}
                </h2>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {t("settings:forms.notification.globalEventsDesc")}
                </p>
              </div>

              <div className="p-4 sm:grid sm:gap-4">
                <EventCheckBoxes />
              </div>
            </div>
          </div>
          {componentMap[type]}

          <FormDebug />
        </div>

        <div className="shrink-0 px-4 border-t border-gray-200 dark:border-gray-700 py-4 sm:px-6">
          <div className="space-x-3 flex justify-end">
            <button
              type="button"
              className="bg-white dark:bg-gray-700 py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-xs text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
              onClick={() => testNotification(form.state.values)}
            >
              {t("settings:forms.notification.test")}
            </button>
            <button
              type="button"
              className="bg-white dark:bg-gray-700 py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-xs text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
              onClick={toggle}
            >
              {t("settings:forms.notification.cancel")}
            </button>
            <button
              type="submit"
              className="inline-flex justify-center py-2 px-4 border border-transparent shadow-xs text-sm font-medium rounded-md text-white bg-blue-600 dark:bg-blue-600 hover:bg-blue-700 dark:hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
            >
              {t("settings:forms.notification.save")}
            </button>
          </div>
        </div>
      </form>
    </form.AppForm>
  );
}

const NotificationTypeSelector = () => {
  const { t } = useTranslation(["options", "settings"]);
  const notificationTypeOptions = getNotificationTypeOptions(t);
  const form = useFormContext();

  return (
    <form.Field name="type">
      {(field) => (
        <>
          <Select
            name={field.name}
            onBlur={field.handleBlur}
            isClearable={true}
            isSearchable={true}
            components={selectComponents}
            placeholder={t("settings:forms.notification.chooseType")}
            styles={selectStyles}
            theme={selectTheme}
            value={field.state.value && notificationTypeOptions.find(o => o.value == field.state.value)}
            onChange={(option: unknown) => {
              // Reset clears the type-specific fields; the shared ones carry over to the new type.
              const { name, enabled, events } = form.state.values as ServiceNotification;
              form.reset();
              form.setFieldValue("name", name, { dontUpdateMeta: true });
              form.setFieldValue("enabled", enabled, { dontUpdateMeta: true });
              form.setFieldValue("events", events, { dontUpdateMeta: true });
              const opt = option as SelectOption | null;
              field.handleChange(opt?.value ?? "");
            }}
            options={notificationTypeOptions}
          />
          <ErrorField meta={field.state.meta} classNames="block text-red-500 mt-2" />
        </>
      )}
    </form.Field>
  );
};

const EventCheckBox = ({ event }: { event: NotificationEventOption; }) => {
  const form = useFormContext();
  const events = useFormValue((v: ServiceNotification) => v.events);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm">
          <span className="font-medium text-gray-900 dark:text-gray-100">{event.label}</span>
          {event.description && <p className="text-gray-500">{event.description}</p>}
        </span>
        <Checkbox
          name={`events-${event.value}`}
          value={events.includes(event.value)}
          setValue={(checked) =>
            form.setFieldValue("events",
              checked
                ? [...events, event.value]
                : events.filter(e => e !== event.value)
            )
          }
        />
      </div>
    </div>
  );
};

const EventCheckBoxes = () => {
  const { t } = useTranslation(["options", "settings"]);
  const type = useFormValue((v: ServiceNotification) => v.type);
  // New releases arrive before filtering, far too many for an inbox meant to be read.
  const eventOptions = getEventOptions(t).filter(e => type !== "BUILTIN" || e.value !== "RELEASE_NEW");

  return (
    <fieldset className="space-y-5">
      <legend className="sr-only">{t("settings:forms.notification.notificationsLegend")}</legend>
      {eventOptions.map((event, idx) => (
        <EventCheckBox
          key={idx}
          event={event}
        />
      ))}
    </fieldset>
  );
};

interface InitialValues {
  id: number;
  enabled: boolean;
  type: NotificationType;
  name: string;
  webhook?: string;
  events: NotificationEvent[];
  method?: string;
  headers?: string;
  used_by_filters?: NotificationFilter[];
}

export function NotificationUpdateForm({ isOpen, toggle, data: notification }: UpdateFormProps<ServiceNotification>) {
  const { t } = useTranslation(["options", "settings"]);
  const isBuiltin = notification.type === "BUILTIN";
  const filterEventOptions: Record<NotificationFilterEvent, string> = {
    "PUSH_APPROVED": t("event.PUSH_APPROVED.label"),
    "PUSH_REJECTED": t("event.PUSH_REJECTED.label"),
    "PUSH_ERROR": t("event.PUSH_ERROR.label"),
    "RELEASE_NEW": t("event.RELEASE_NEW.label"),
  };
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (notification: ServiceNotification) => APIClient.notifications.update(notification),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NotificationKeys.lists() });

      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.notification.updated", { name: notification.name })} t={toastInstance} />);
      toggle();
    }
  });

  const onSubmit = (formData: InitialValues) => mutation.mutate(formData);

  const deleteMutation = useMutation({
    mutationFn: (notificationID: number) => APIClient.notifications.delete(notificationID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NotificationKeys.lists() });

      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.notification.deleted", { name: notification.name })} t={toastInstance} />);
    }
  });

  const deleteAction = () => deleteMutation.mutate(notification.id);

  const testMutation = useMutation({
    mutationFn: (n: ServiceNotification) => APIClient.notifications.test(n)
  });

  const testNotification = (data: unknown) => testMutation.mutate(data as ServiceNotification);

  const initialValues: InitialValues = {
    id: notification.id,
    enabled: notification.enabled,
    type: notification.type,
    name: notification.name,
    webhook: notification.webhook,
    events: notification.events || [],
    method: notification.method,
    headers: notification.headers,
    used_by_filters: notification.used_by_filters || [],
  };

  return (
    <SlideOver<InitialValues>
      type="UPDATE"
      title={t("settings:forms.notification.title")}
      isOpen={isOpen}
      toggle={toggle}
      onSubmit={onSubmit}
      deleteAction={isBuiltin ? undefined : deleteAction}
      initialValues={initialValues}
      validate={(values) => validateNotification(values, t("settings:forms.notification.required"))}
      testFn={testNotification}
    >
      {(values) => (
        <div>
          <TextFieldWide name="name" label={t("settings:forms.notification.name")} required={true} />

          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {!isBuiltin && (
              <div className="py-4 flex items-center justify-between space-y-1 px-4 sm:space-y-0 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-4">
                <div>
                  <label
                    htmlFor="type"
                    className="block text-sm font-medium text-gray-900 dark:text-white"
                  >
                    {t("settings:forms.notification.type")}
                  </label>
                </div>
                <div className="sm:col-span-2">
                  <NotificationTypeSelector />
                </div>
              </div>
            )}
            <SwitchGroupWide name="enabled" label={t("settings:forms.notification.enabled")} />
            <div className="pb-2">
              <div className="p-4">
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                  {t("settings:forms.notification.globalEvents")}
                </h2>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {t("settings:forms.notification.globalEventsDesc")}
                </p>
              </div>

              <div className="p-4 sm:grid sm:gap-4">
                <EventCheckBoxes />
              </div>
            </div>
          </div>

          <div className="pb-2">
            <div className="p-4">
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">
                  {t("settings:forms.notification.perFilterEvents")}
                </h2>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {t("settings:forms.notification.perFilterEventsDesc")}
                </p>
            </div>

            <div className="p-4 sm:grid sm:gap-4">
              {values.used_by_filters && values.used_by_filters?.length > 0
                ? values.used_by_filters?.map(f => (
                  <Link key={f.filter_id} to="/filters/$filterId/notifications" params={{ filterId: f.filter_id }}>
                    <div key={f.filter_id} className="flex justify-between px-2 py-2 bg-gray-50 dark:bg-gray-750 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md">
                      <span className="font-medium text-gray-500 dark:text-gray-300">{f.filter_name}</span>
                      <div className="flex gap-2">
                        {f.events.length > 0
                          ? f.events.map((e) => <span className="inline-flex items-center rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 dark:bg-gray-400/10 dark:text-gray-400">{filterEventOptions[e]}</span>)
                          : <span className="inline-flex items-center rounded-md bg-yellow-100 px-2 py-1 text-xs font-medium text-yellow-600 dark:bg-yellow-400/10 dark:text-yellow-400">{t("settings:forms.notification.muted")}</span>}
                      </div>
                    </div>
                  </Link>
                ))
                :
                <EmptySimple
                  title={t("settings:forms.notification.notUsedInFilters")}
                  subtitle=""
                  border={true}
                />

              }
            </div>
          </div>

          {componentMap[values.type]}

        </div>
      )}
    </SlideOver>
  );
}
