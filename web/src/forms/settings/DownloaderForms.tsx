/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useRef, useState, ReactElement } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { XMarkIcon } from "@heroicons/react/24/solid";
import { useSelector } from "@tanstack/react-form";
import { useTranslation } from "react-i18next";

import { classNames, sleep } from "@utils";
import { FormDebug } from "@components/debug";
import { APIClient } from "@api/APIClient";
import { DownloaderKeys } from "@api/query_keys";
import { getDownloaderTypeOptions } from "@domain/constants";
import { toast } from "@components/hot-toast";
import Toast from "@components/notifications/Toast";
import { useToggle } from "@hooks/hooks";
import { useAppForm, useFormValue } from "@hooks/form";
import { DeleteModal } from "@components/modals";
import { SlideOverShell, SlideOverTitle } from "@components/panels";
import {
  NumberFieldWide,
  PasswordFieldWide,
  RadioFieldsetWide,
  SwitchGroupWide,
  TextFieldWide
} from "@components/inputs";
import { DocsLink } from "@components/ExternalLink";
import { AddFormProps, UpdateFormProps } from "@forms/_shared";

interface InitialValuesSettings {
  basic?: {
    auth: boolean;
    username: string;
    password: string;
  };
  auth?: {
    enabled: boolean;
    type: string;
    username: string;
    password: string;
  };
  rules?: {
    enabled?: boolean;
    ignore_slow_torrents?: boolean;
    ignore_slow_torrents_condition?: IgnoreTorrentsCondition;
    download_speed_threshold?: number;
    max_active_downloads?: number;
  };
  external_download_client_id?: number;
  external_download_client?: string;
}

interface InitialValues {
  name: string;
  type: DownloaderType;
  enabled: boolean;
  host: string;
  port: number;
  tls: boolean;
  tls_skip_verify: boolean;
  username: string;
  password: string;
  settings: InitialValuesSettings;
}

function FormFieldsArr() {
  const { t } = useTranslation("settings");
  const { tls, settings } = useFormValue((v: InitialValues) => ({ tls: v.tls, settings: v.settings }));

  return (
    <div className="flex flex-col space-y-4 px-1 mb-4 sm:py-0 sm:space-y-0">
      <TextFieldWide
        required
        name="host"
        label={t("forms.downloadClient.host")}
        help={t("forms.downloadClient.hostHelpArr")}
        tooltip={
          <div>
            <p>{t("forms.downloadClient.guidesArr")}</p>
            <br />
            <p>{t("forms.downloadClient.dedicatedServers")}</p>
            <DocsLink href="https://autobrr.com/configuration/download-clients/dedicated/#sonarr" />
            <p>{t("forms.downloadClient.sharedSeedboxProviders")}</p>
            <DocsLink href="https://autobrr.com/configuration/download-clients/shared-seedboxes#sonarr" />
          </div>
        }
      />

      <SwitchGroupWide name="tls" label={t("forms.downloadClient.tls")} />

      {tls && (
        <SwitchGroupWide
          name="tls_skip_verify"
          label={t("forms.downloadClient.skipTls")}
        />
      )}

      <PasswordFieldWide required name="settings.apikey" label={t("forms.downloadClient.apiKey")} />

      <SwitchGroupWide name="settings.basic.auth" label={t("forms.downloadClient.basicAuth")} />

      {settings.basic?.auth === true && (
        <>
          <TextFieldWide name="settings.basic.username" label={t("forms.downloadClient.username")} />
          <PasswordFieldWide name="settings.basic.password" label={t("forms.downloadClient.password")} />
        </>
      )}
    </div>
  );
}

export interface componentMapType {
  [key: string]: ReactElement;
}

const componentMap: componentMapType = {
  RADARR: <FormFieldsArr />,
  SONARR: <FormFieldsArr />,
  LIDARR: <FormFieldsArr />,
  WHISPARR: <FormFieldsArr />,
  WHISPARR_V3: <FormFieldsArr />,
  READARR: <FormFieldsArr />,
  SPORTARR: <FormFieldsArr />,
};

function FormFieldsRulesArr() {
  const { t } = useTranslation("settings");

  return (
    <div className="border-t border-gray-200 dark:border-gray-700 py-5 px-2">
      <div className="px-4">
        <h2 className="text-lg font-medium text-gray-900 dark:text-white">
          {t("forms.downloadClient.downloadClientTitle")}
        </h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {t("forms.downloadClient.downloadClientDescription")}
        </p>
      </div>

      <TextFieldWide name="settings.external_download_client" label={t("forms.downloadClient.clientName")} tooltip={<div><p>{t("forms.downloadClient.clientNameTooltip")}</p></div>} />

      <NumberFieldWide name="settings.external_download_client_id" label={t("forms.downloadClient.clientIdDeprecated")} tooltip={<div><p>{t("forms.downloadClient.clientIdDeprecatedTooltip")}</p></div>} />
    </div>
  );
}

const rulesComponentMap: componentMapType = {
  RADARR: <FormFieldsRulesArr />,
  SONARR: <FormFieldsRulesArr />,
  LIDARR: <FormFieldsRulesArr />,
  WHISPARR: <FormFieldsRulesArr />,
  WHISPARR_V3: <FormFieldsRulesArr />,
  READARR: <FormFieldsRulesArr />,
  SPORTARR: <FormFieldsRulesArr />,
};

interface formButtonsProps {
  isSuccessfulTest: boolean;
  isErrorTest: boolean;
  isTesting: boolean;
  cancelFn: () => void;
  testFn: () => void;
  type: "CREATE" | "UPDATE";
  toggleDeleteModal?: () => void;
}

function DownloaderFormButtons({
  type,
  isSuccessfulTest,
  isErrorTest,
  isTesting,
  cancelFn,
  testFn,
  toggleDeleteModal
}: formButtonsProps) {
  const { t } = useTranslation("settings");

  return (
    <div className="shrink-0 px-4 border-t border-gray-200 dark:border-gray-700 py-5 sm:px-6">
      <div className={classNames(type === "CREATE" ? "justify-end" : "justify-between", "space-x-3 flex")}>
        {type === "UPDATE" && (
          <button
            type="button"
            className="inline-flex items-center justify-center px-4 py-2 border border-transparent font-medium rounded-md text-red-700 dark:text-white bg-red-100 dark:bg-red-700 hover:bg-red-200 dark:hover:bg-red-600 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:text-sm cursor-pointer"
            onClick={toggleDeleteModal}
          >
            {t("forms.downloadClient.remove")}
          </button>
        )}
        <div className="flex">
          <button
            type="button"
            className={classNames(
              isSuccessfulTest
                ? "text-green-500 border-green-500 bg-green-50"
                : isErrorTest
                  ? "text-red-500 border-red-500 bg-red-50"
                  : "border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-400 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:border-rose-700 active:bg-rose-700",
              isTesting ? "cursor-not-allowed" : "cursor-pointer",
              "mr-2 inline-flex items-center px-4 py-2 border font-medium rounded-md shadow-xs text-sm transition ease-in-out duration-150 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500"
            )}
            disabled={isTesting}
            // onClick={() => testClient(values)}
            onClick={testFn}
          >
            {isTesting ? (
              <svg
                className="animate-spin h-5 w-5 text-green-500"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                ></circle>
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
            ) : isSuccessfulTest ? (
              t("forms.downloadClient.ok")
            ) : isErrorTest ? (
              t("forms.downloadClient.error")
            ) : (
              t("forms.downloadClient.test")
            )}
          </button>

          <button
            type="button"
            className="mr-4 bg-white dark:bg-gray-700 py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-xs text-sm font-medium text-gray-700 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
            onClick={cancelFn}
          >
            {t("forms.downloadClient.cancel")}
          </button>
          <button
            type="submit"
            className="inline-flex justify-center py-2 px-4 border border-transparent shadow-xs text-sm font-medium rounded-md text-white bg-blue-600 dark:bg-blue-600 hover:bg-blue-700 dark:hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
          >
            {type === "CREATE" ? t("forms.downloadClient.create") : t("forms.downloadClient.save")}
          </button>
        </div>
      </div>
    </div>
  );
}

export function DownloaderAddForm({ isOpen, toggle }: AddFormProps) {
  return (
    <SlideOverShell isOpen={isOpen} toggle={toggle}>
      <DownloaderAddFormPanel toggle={toggle} />
    </SlideOverShell>
  );
}

interface DownloaderAddFormPanelProps {
  toggle: () => void;
}

function DownloaderAddFormPanel({ toggle }: DownloaderAddFormPanelProps) {
  const { t } = useTranslation(["options", "settings"]);
  const downloaderTypeOptions = getDownloaderTypeOptions(t);
  const [isTesting, setIsTesting] = useState(false);
  const [isSuccessfulTest, setIsSuccessfulTest] = useState(false);
  const [isErrorTest, setIsErrorTest] = useState(false);

  const queryClient = useQueryClient();

  const addMutation = useMutation({
    mutationFn: (client: Downloader) => APIClient.downloaders.create(client),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DownloaderKeys.lists() });
      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.downloadClient.added")} t={toastInstance} />);

      toggle();
    },
    onError: () => {
      toast.custom((toastInstance) => <Toast type="error" body={t("settings:forms.downloadClient.addFailed")} t={toastInstance} />);
    }
  });

  const onSubmit = (data: unknown) => addMutation.mutate(data as Downloader);

  const testClientMutation = useMutation({
    mutationFn: (client: Downloader) => APIClient.downloaders.test(client),
    onMutate: () => {
      setIsTesting(true);
      setIsErrorTest(false);
      setIsSuccessfulTest(false);
    },
    onSuccess: () => {
      sleep(1000)
        .then(() => {
          setIsTesting(false);
          setIsSuccessfulTest(true);
        })
        .then(() => {
          sleep(2500).then(() => {
            setIsSuccessfulTest(false);
          });
        });
    },
    onError: () => {
      setIsTesting(false);
      setIsErrorTest(true);
      sleep(2500).then(() => {
        setIsErrorTest(false);
      });
    }
  });

  const testClient = (data: unknown) => testClientMutation.mutate(data as Downloader);

  const initialValues: InitialValues = {
    name: "",
    type: "SONARR",
    enabled: true,
    host: "",
    port: 0,
    tls: false,
    tls_skip_verify: false,
    username: "",
    password: "",
    settings: {}
  };

  const form = useAppForm({
    defaultValues: initialValues,
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
                  {t("settings:forms.downloadClient.addTitle")}
                </SlideOverTitle>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {t("settings:forms.downloadClient.addDescription")}
                </p>
              </div>
              <div className="h-7 flex items-center">
                <button
                  type="button"
                  className="bg-white dark:bg-gray-800 rounded-md text-gray-400 hover:text-gray-500 focus:outline-hidden focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
                  onClick={toggle}
                >
                  <span className="sr-only">{t("settings:forms.downloadClient.closePanel")}</span>
                  <XMarkIcon
                    className="h-6 w-6"
                    aria-hidden="true"
                  />
                </button>
              </div>
            </div>
          </div>

          <div className="flex flex-col space-y-4 px-1 py-6 sm:py-0 sm:space-y-0">
            <TextFieldWide required name="name" label={t("settings:forms.downloadClient.name")} />
            <SwitchGroupWide name="enabled" label={t("settings:forms.downloadClient.enabled")} />
            <RadioFieldsetWide
              name="type"
              legend={t("settings:forms.downloadClient.type")}
              options={downloaderTypeOptions}
            />
            <div>{componentMap[type]}</div>
          </div>

          {rulesComponentMap[type]}

          <FormDebug />
        </div>

        <DownloaderFormButtons
          type="CREATE"
          isTesting={isTesting}
          isSuccessfulTest={isSuccessfulTest}
          isErrorTest={isErrorTest}
          cancelFn={toggle}
          testFn={() => testClient(form.state.values)}
        />
      </form>
    </form.AppForm>
  );
}

export function DownloaderUpdateForm({ isOpen, toggle, data }: UpdateFormProps<Downloader>) {
  const cancelButtonRef = useRef(null);

  return (
    <SlideOverShell isOpen={isOpen} toggle={toggle} initialFocus={cancelButtonRef}>
      <DownloaderUpdateFormPanel toggle={toggle} data={data} />
    </SlideOverShell>
  );
}

interface DownloaderUpdateFormPanelProps {
  toggle: () => void;
  data: Downloader;
}

function DownloaderUpdateFormPanel({ toggle, data: client }: DownloaderUpdateFormPanelProps) {
  const { t } = useTranslation(["options", "settings"]);
  const downloaderTypeOptions = getDownloaderTypeOptions(t);
  const [isTesting, setIsTesting] = useState(false);
  const [isSuccessfulTest, setIsSuccessfulTest] = useState(false);
  const [isErrorTest, setIsErrorTest] = useState(false);
  const [deleteModalIsOpen, toggleDeleteModal] = useToggle(false);

  const cancelModalButtonRef = useRef(null);

  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (client: Downloader) => APIClient.downloaders.update(client),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DownloaderKeys.lists() });
      queryClient.invalidateQueries({ queryKey: DownloaderKeys.detail(client.id) });

      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.downloadClient.updated", { name: client.name })} t={toastInstance} />);
      toggle();
    }
  });

  const onSubmit = (data: Downloader) => mutation.mutate(data);

  const deleteMutation = useMutation({
    mutationFn: (clientID: number) => APIClient.downloaders.delete(clientID),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DownloaderKeys.lists() });
      queryClient.invalidateQueries({ queryKey: DownloaderKeys.detail(client.id) });

      toast.custom((toastInstance) => <Toast type="success" body={t("settings:forms.downloadClient.deleted", { name: client.name })} t={toastInstance} />);
      toggleDeleteModal();
    }
  });

  const deleteAction = () => deleteMutation.mutate(client.id);


  const testClientMutation = useMutation({
    mutationFn: (client: Downloader) => APIClient.downloaders.test(client),
    onMutate: () => {
      setIsTesting(true);
      setIsErrorTest(false);
      setIsSuccessfulTest(false);
    },
    onSuccess: () => {
      sleep(1000)
        .then(() => {
          setIsTesting(false);
          setIsSuccessfulTest(true);
        })
        .then(() => {
          sleep(2500).then(() => {
            setIsSuccessfulTest(false);
          });
        });
    },
    onError: () => {
      setIsTesting(false);
      setIsErrorTest(true);
      sleep(2500).then(() => {
        setIsErrorTest(false);
      });
    }
  });

  const testClient = (data: unknown) => testClientMutation.mutate(data as Downloader);

  const initialValues = {
    id: client.id,
    name: client.name,
    type: client.type,
    enabled: client.enabled,
    host: client.host,
    port: client.port,
    tls: client.tls,
    tls_skip_verify: client.tls_skip_verify,
    username: client.username,
    password: client.password,
    settings: client.settings
  };

  const form = useAppForm({
    defaultValues: initialValues,
    onSubmit: ({ value }) => onSubmit(value)
  });

  const type = useSelector(form.store, (state) => state.values.type);

  return (
    <>
      <DeleteModal
        isOpen={deleteModalIsOpen}
        isLoading={deleteMutation.isPending}
        toggle={toggleDeleteModal}
        buttonRef={cancelModalButtonRef}
        deleteAction={deleteAction}
        title={t("settings:forms.downloadClient.removeTitle")}
        text={t("settings:forms.downloadClient.removeText")}
      />
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
                    {t("settings:forms.downloadClient.editTitle")}
                  </SlideOverTitle>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {t("settings:forms.downloadClient.editDescription")}
                  </p>
                </div>
                <div className="h-7 flex items-center">
                  <button
                    type="button"
                    className="bg-white dark:bg-gray-800 rounded-md text-gray-400 hover:text-gray-500 focus:outline-hidden focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
                    onClick={toggle}
                  >
                    <span className="sr-only">{t("settings:forms.downloadClient.closePanel")}</span>
                    <XMarkIcon
                      className="h-6 w-6"
                      aria-hidden="true"
                    />
                  </button>
                </div>
              </div>
            </div>

            <div className="py-6 space-y-6 sm:py-0 sm:space-y-0 sm:divide-y dark:divide-gray-700">
              <TextFieldWide required name="name" label={t("settings:forms.downloadClient.name")} />
              <SwitchGroupWide name="enabled" label={t("settings:forms.downloadClient.enabled")} />
              <RadioFieldsetWide
                name="type"
                legend={t("settings:forms.downloadClient.type")}
                options={downloaderTypeOptions}
              />
              <div>{componentMap[type]}</div>
            </div>

            {rulesComponentMap[type]}

            <FormDebug />
          </div>

          <DownloaderFormButtons
            type="UPDATE"
            toggleDeleteModal={toggleDeleteModal}
            isTesting={isTesting}
            isSuccessfulTest={isSuccessfulTest}
            isErrorTest={isErrorTest}
            cancelFn={toggle}
            testFn={() => testClient(form.state.values)}
          />
        </form>
      </form.AppForm>
    </>
  );
}
